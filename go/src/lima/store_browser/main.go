package store_browser

import (
	"bufio"
	"net/url"
	"sync"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/toml"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/kilo/sku_fmt"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

const DefaultTimeout = 2e9

type transacted struct {
	sync.Mutex
	interfaces.MutableSetLike[*ids.ObjectId]
}

type checkedOutWithItem struct {
	*sku.CheckedOut
	Item
}

type Store struct {
	config            store_config.Store
	externalStoreInfo store_workspace.Supplies
	typ               ids.Type
	browser           browser_items.BrowserProxy

	tabCache cache

	urls map[url.URL][]Item

	l       sync.Mutex
	deleted map[url.URL][]checkedOutWithItem
	added   map[url.URL][]checkedOutWithItem

	itemsById map[string]Item

	transacted transacted

	transactedUrlIndex  map[url.URL]sku.TransactedMutableSet
	transactedItemIndex map[browser_items.ItemId]*sku.Transacted

	itemDeletedStringFormatWriter interfaces.FuncIter[*sku.CheckedOut]
}

func Make(
	k store_config.Store,
	s env_repo.Env,
	itemDeletedStringFormatWriter interfaces.FuncIter[*sku.CheckedOut],
) *Store {
	c := &Store{
		config:    k,
		typ:       ids.MustType("toml-bookmark"),
		deleted:   make(map[url.URL][]checkedOutWithItem),
		added:     make(map[url.URL][]checkedOutWithItem),
		itemsById: make(map[string]Item),
		transacted: transacted{
			MutableSetLike: collections_value.MakeMutableValueSet(
				quiter.StringerKeyer[*ids.ObjectId]{},
			),
		},
		transactedUrlIndex: make(
			map[url.URL]sku.TransactedMutableSet,
		),
		transactedItemIndex: make(
			map[browser_items.ItemId]*sku.Transacted,
		),
		itemDeletedStringFormatWriter: itemDeletedStringFormatWriter,
	}

	return c
}

func (fs *Store) GetExternalStoreLike() store_workspace.StoreLike {
	return fs
}

func (s *Store) ReadAllExternalItems() error {
	return nil
}

func (s *Store) GetObjectIdsForString(
	v string,
) (k []sku.ExternalObjectId, err error) {
	item, ok := s.itemsById[v]

	if !ok {
		err = errors.ErrorWithStackf("not a browser item id")
		return
	}

	k = append(k, item.GetExternalObjectId())

	return
}

func (s *Store) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	wg.Do(s.flushUrls)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO limit this to being used only by *Item.ReadFromExternal
func (c *Store) getUrl(sk *sku.Transacted) (u *url.URL, err error) {
	var r interfaces.ReadCloseDigester

	if r, err = c.externalStoreInfo.GetDefaultBlobStore().BlobReader(sk.GetBlobSha()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, r)

	var tb sku_fmt.TomlBookmark

	dec := toml.NewDecoder(r)

	if err = dec.Decode(&tb); err != nil {
		err = errors.Wrapf(
			err,
			"Sha: %s, Object Id: %s",
			sk.GetBlobSha(),
			sk.GetObjectId(),
		)
		return
	}

	if u, err = url.Parse(tb.Url); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (c *Store) CheckoutOne(
	options checkout_options.Options,
	tg sku.TransactedGetter,
) (cz sku.SkuType, err error) {
	sz := tg.GetSku()

	if !sz.Metadata.Type.Equals(c.typ) {
		err = errors.Wrap(env_workspace.ErrUnsupportedType(sz.Metadata.Type))
		return
	}

	var u *url.URL

	if u, err = c.getUrl(sz); err != nil {
		err = errors.Wrap(err)
		return
	}

	co := GetCheckedOutPool().Get()
	cz = co
	var item Item

	if err = item.Url.Set(u.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	item.ExternalId = sz.ObjectId.String()
	item.Id.Type = "tab"

	sku.TransactedResetter.ResetWith(co.GetSku(), sz)
	sku.TransactedResetter.ResetWith(co.GetSkuExternal().GetSku(), sz)
	co.SetState(checked_out_state.JustCheckedOut)
	co.GetSkuExternal().ExternalType = ids.MustType("!browser-tab")

	if err = item.WriteToExternal(co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return
	}

	co.GetSkuExternal().RepoId = c.externalStoreInfo.RepoId

	c.l.Lock()
	defer c.l.Unlock()

	existing := c.added[*u]
	c.added[*u] = append(existing, checkedOutWithItem{
		CheckedOut: co.Clone(),
		Item:       item,
	})

	return
}

func (c *Store) QueryCheckedOut(
	qg *query.Query,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	// o := sku.CommitOptions{
	// 	Mode: object_mode.ModeRealizeSansProto,
	// }

	ex := executor{
		store: c,
		qg:    qg,
		out:   f,
	}

	for u, items := range c.urls {
		matchingUrls, exactIndexURLMatch := c.transactedUrlIndex[u]

		for _, item := range items {
			var matchingTabId *sku.Transacted
			var trackedFromBefore bool

			tabId := item.Id
			matchingTabId, trackedFromBefore = c.transactedItemIndex[tabId]

			if trackedFromBefore {
				if err = ex.tryToEmitOneExplicitlyCheckedOut(
					matchingTabId,
					item,
				); err != nil {
					err = errors.Wrapf(err, "Item: %#v", item)
					return
				}
			} else if !exactIndexURLMatch {
				if err = ex.tryToEmitOneUntracked(item); err != nil {
					err = errors.Wrapf(err, "Item: %#v", item)
					return
				}
			} else if exactIndexURLMatch {
				if err = matchingUrls.Each(
					func(matching *sku.Transacted) (err error) {
						if err = ex.tryToEmitOneRecognized(
							matching,
							item,
						); err != nil {
							err = errors.Wrapf(err, "Item: %#v", item)
							return
						}

						return
					},
				); err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		}
	}

	return
}

// TODO support updating bookmarks without overwriting. Maybe move to
// toml-bookmark type
func (s *Store) SaveBlob(e sku.ExternalLike) (err error) {
	var blobWriter interfaces.WriteCloseDigester

	if blobWriter, err = s.externalStoreInfo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobWriter)

	var item Item

	if err = item.ReadFromExternal(e.GetSku()); err != nil {
		err = errors.Wrap(err)
		return
	}

	tb := sku_fmt.TomlBookmark{
		Url: item.Url.String(),
	}

	func() {
		bw := bufio.NewWriter(blobWriter)
		defer errors.DeferredFlusher(&err, bw)

		enc := toml.NewEncoder(bw)

		if err = enc.Encode(tb); err != nil {
			err = errors.Wrap(err)
			return
		}
	}()

	e.GetSku().Metadata.Blob.SetDigester(blobWriter)

	return
}

func (s *Store) asBlobSaver() sku.BlobSaver {
	return s
}

func (s *Store) UpdateCheckoutFromCheckedOut(
	options checkout_options.OptionsWithoutMode,
	col sku.SkuType,
) (err error) {
	return
}
