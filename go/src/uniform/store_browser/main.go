package store_browser

import (
	"bufio"
	"net/url"
	"sync"

	"code.linenisgreat.com/chrest/go/src/charlie/browser_items"
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/charlie/toml"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/mike/sku_json_fmt"
	"code.linenisgreat.com/dodder/go/src/papa/queries"
	"code.linenisgreat.com/dodder/go/src/quebec/store_workspace"
	"code.linenisgreat.com/dodder/go/src/sierra/env_workspace"
	"code.linenisgreat.com/dodder/go/src/tango/store_config"
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

func (store *Store) GetExternalStoreLike() store_workspace.StoreLike {
	return store
}

func (store *Store) ReadAllExternalItems() error {
	return nil
}

func (store *Store) GetObjectIdsForString(
	v string,
) (k []sku.ExternalObjectId, err error) {
	item, ok := store.itemsById[v]

	if !ok {
		err = errors.ErrorWithStackf("not a browser item id")
		return k, err
	}

	k = append(k, item.GetExternalObjectId())

	return k, err
}

func (store *Store) Flush() (err error) {
	wg := errors.MakeWaitGroupParallel()

	wg.Do(store.flushUrls)

	if err = wg.GetError(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO limit this to being used only by *Item.ReadFromExternal
func (store *Store) getUrl(sk *sku.Transacted) (u *url.URL, err error) {
	var r interfaces.BlobReader

	if r, err = store.externalStoreInfo.GetDefaultBlobStore().MakeBlobReader(sk.GetBlobDigest()); err != nil {
		err = errors.Wrap(err)
		return u, err
	}

	defer errors.DeferredCloser(&err, r)

	var tomlBookmark sku_json_fmt.TomlBookmark

	dec := toml.NewDecoder(r)

	if err = dec.Decode(&tomlBookmark); err != nil {
		err = errors.Wrapf(
			err,
			"Sha: %s, Object Id: %s",
			sk.GetBlobDigest(),
			sk.GetObjectId(),
		)
		return u, err
	}

	if u, err = url.Parse(tomlBookmark.Url); err != nil {
		err = errors.Wrap(err)
		return u, err
	}

	return u, err
}

func (store *Store) CheckoutOne(
	options checkout_options.Options,
	tg sku.TransactedGetter,
) (cz sku.SkuType, err error) {
	sz := tg.GetSku()

	if !sz.GetMetadata().GetType().Equals(store.typ) {
		err = errors.Wrap(env_workspace.ErrUnsupportedType(sz.GetMetadata().GetType()))
		return cz, err
	}

	var u *url.URL

	if u, err = store.getUrl(sz); err != nil {
		err = errors.Wrap(err)
		return cz, err
	}

	co := GetCheckedOutPool().Get()
	cz = co
	var item Item

	if err = item.Url.Set(u.String()); err != nil {
		err = errors.Wrap(err)
		return cz, err
	}

	item.ExternalId = sz.ObjectId.String()
	item.Id.Type = "tab"

	sku.TransactedResetter.ResetWith(co.GetSku(), sz)
	sku.TransactedResetter.ResetWith(co.GetSkuExternal().GetSku(), sz)
	co.SetState(checked_out_state.JustCheckedOut)
	co.GetSkuExternal().ExternalType = ids.MustType("!browser-tab")

	if err = item.WriteToExternal(co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return cz, err
	}

	co.GetSkuExternal().RepoId = store.externalStoreInfo.RepoId

	store.l.Lock()
	defer store.l.Unlock()

	existing := store.added[*u]
	store.added[*u] = append(existing, checkedOutWithItem{
		CheckedOut: co.Clone(),
		Item:       item,
	})

	return cz, err
}

func (store *Store) QueryCheckedOut(
	qg *queries.Query,
	f interfaces.FuncIter[sku.SkuType],
) (err error) {
	// o := sku.CommitOptions{
	// 	Mode: object_mode.ModeRealizeSansProto,
	// }

	ex := executor{
		store: store,
		qg:    qg,
		out:   f,
	}

	for u, items := range store.urls {
		matchingUrls, exactIndexURLMatch := store.transactedUrlIndex[u]

		for _, item := range items {
			var matchingTabId *sku.Transacted
			var trackedFromBefore bool

			tabId := item.Id
			matchingTabId, trackedFromBefore = store.transactedItemIndex[tabId]

			if trackedFromBefore {
				if err = ex.tryToEmitOneExplicitlyCheckedOut(
					matchingTabId,
					item,
				); err != nil {
					err = errors.Wrapf(err, "Item: %#v", item)
					return err
				}
			} else if !exactIndexURLMatch {
				if err = ex.tryToEmitOneUntracked(item); err != nil {
					err = errors.Wrapf(err, "Item: %#v", item)
					return err
				}
			} else if exactIndexURLMatch {
				for matching := range matchingUrls.All() {
					if err = ex.tryToEmitOneRecognized(
						matching,
						item,
					); err != nil {
						err = errors.Wrapf(err, "Item: %#v", item)
						return err
					}
				}
			}
		}
	}

	return err
}

// TODO support updating bookmarks without overwriting. Maybe move to
// toml-bookmark type
func (store *Store) SaveBlob(object sku.ExternalLike) (err error) {
	var blobWriter interfaces.BlobWriter

	if blobWriter, err = store.externalStoreInfo.GetDefaultBlobStore().MakeBlobWriter(
		nil,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, blobWriter)

	var item Item

	if err = item.ReadFromExternal(object.GetSku()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	tomlBookmark := sku_json_fmt.TomlBookmark{
		Url: item.Url.String(),
	}

	func() {
		bw := bufio.NewWriter(blobWriter)
		defer errors.DeferredFlusher(&err, bw)

		enc := toml.NewEncoder(bw)

		if err = enc.Encode(tomlBookmark); err != nil {
			err = errors.Wrap(err)
			return
		}
	}()

	markl.SetDigester(
		object.GetSku().Metadata.GetBlobDigestMutable(),
		blobWriter,
	)

	return err
}

func (store *Store) asBlobSaver() sku.BlobSaver {
	return store
}

func (store *Store) UpdateCheckoutFromCheckedOut(
	options checkout_options.OptionsWithoutMode,
	col sku.SkuType,
) (err error) {
	return err
}
