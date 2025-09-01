package store_abbr

import (
	"bufio"
	"encoding/gob"
	"io"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/charlie/tridex"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type indexAbbrEncodableTridexes struct {
	BlobId   indexNotZettelId[markl.Id, *markl.Id]
	ZettelId indexZettelId
}

type indexAbbr struct {
	options_print.Options

	lock    sync.Locker
	once    *sync.Once
	envRepo env_repo.Env

	path string

	indexAbbrEncodableTridexes

	didRead    bool
	hasChanges bool
}

func NewIndexAbbr(
	options options_print.Options,
	envRepo env_repo.Env,
) (i *indexAbbr, err error) {
	i = &indexAbbr{
		Options: options,
		lock:    &sync.Mutex{},
		once:    &sync.Once{},
		path:    envRepo.DirCache("Abbr"),
		envRepo: envRepo,
		indexAbbrEncodableTridexes: indexAbbrEncodableTridexes{
			BlobId: indexNotZettelId[markl.Id, *markl.Id]{
				ObjectIds: tridex.Make(),
			},
			ZettelId: indexZettelId{
				Heads: tridex.Make(),
				Tails: tridex.Make(),
			},
		},
	}

	i.indexAbbrEncodableTridexes.ZettelId.readFunc = i.readIfNecessary
	i.indexAbbrEncodableTridexes.BlobId.readFunc = i.readIfNecessary

	return
}

func (index *indexAbbr) Flush() (err error) {
	index.lock.Lock()
	defer index.lock.Unlock()

	if !index.hasChanges {
		ui.Log().Print("no changes")
		return
	}

	var w1 io.WriteCloser

	if w1, err = index.envRepo.WriteCloserCache(index.path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, w1)

	w := bufio.NewWriter(w1)

	defer errors.DeferredFlusher(&err, w)

	enc := gob.NewEncoder(w)

	if err = enc.Encode(index.indexAbbrEncodableTridexes); err != nil {
		err = errors.Wrapf(err, "failed to write encoded object id")
		return
	}

	return
}

func (index *indexAbbr) readIfNecessary() (err error) {
	index.once.Do(
		func() {
			if index.didRead {
				return
			}

			ui.Log().Print("reading")

			index.didRead = true

			var r1 io.ReadCloser

			if r1, err = index.envRepo.ReadCloserCache(index.path); err != nil {
				if errors.IsNotExist(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
				}

				return
			}

			defer errors.Deferred(&err, r1.Close)

			r := bufio.NewReader(r1)

			dec := gob.NewDecoder(r)

			ui.Log().Print("starting decode")

			if err = dec.Decode(&index.indexAbbrEncodableTridexes); err != nil {
				ui.Log().Print("finished decode unsuccessfully")
				err = errors.Wrap(err)
				return
			}
		},
	)

	return
}

func (index *indexAbbr) GetAbbr() (out ids.Abbr) {
	out.ZettelId.Expand = index.ZettelId().ExpandStringString
	out.BlobId.Expand = index.BlobId().ExpandStringString

	if index.AbbreviateZettelIds {
		out.ZettelId.Abbreviate = index.ZettelId().Abbreviate
	}

	if index.AbbreviateMarklIds {
		out.BlobId.Abbreviate = index.BlobId().Abbreviate
	}

	return
}

func (index *indexAbbr) AddObjectToAbbreviationStore(
	object *sku.Transacted,
) (err error) {
	if err = index.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	index.hasChanges = true

	index.indexAbbrEncodableTridexes.BlobId.ObjectIds.Add(
		markl.Format(object.GetBlobDigest()),
	)

	objectIdString := object.GetObjectId().String()

	switch object.GetGenre() {
	case genres.Zettel:
		var zettelId ids.ZettelId

		if err = zettelId.SetFromIdParts(object.GetObjectId().Parts()); err != nil {
			err = errors.Wrap(err)
			return
		}

		index.indexAbbrEncodableTridexes.ZettelId.Heads.Add(zettelId.GetHead())
		index.indexAbbrEncodableTridexes.ZettelId.Tails.Add(zettelId.GetTail())

	case genres.Type,
		genres.Tag,
		genres.Config,
		genres.InventoryList,
		genres.Repo:
		return

	default:
		err = errors.ErrorWithStackf(
			"unsupported object id: %#v",
			objectIdString,
		)
		return
	}

	return
}

func (index *indexAbbr) ZettelId() sku.AbbrStoreGeneric[ids.ZettelId, *ids.ZettelId] {
	return &index.indexAbbrEncodableTridexes.ZettelId
}

func (index *indexAbbr) BlobId() sku.AbbrStoreGeneric[markl.Id, *markl.Id] {
	return &index.indexAbbrEncodableTridexes.BlobId
}
