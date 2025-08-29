package store_abbr

import (
	"bufio"
	"encoding/gob"
	"io"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/charlie/tridex"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type indexAbbrEncodableTridexes struct {
	BlobId   indexNotZettelId[merkle.Id, *merkle.Id]
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
			BlobId: indexNotZettelId[merkle.Id, *merkle.Id]{
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

func (i *indexAbbr) Flush() (err error) {
	i.lock.Lock()
	defer i.lock.Unlock()

	if !i.hasChanges {
		ui.Log().Print("no changes")
		return
	}

	var w1 io.WriteCloser

	if w1, err = i.envRepo.WriteCloserCache(i.path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, w1)

	w := bufio.NewWriter(w1)

	defer errors.DeferredFlusher(&err, w)

	enc := gob.NewEncoder(w)

	if err = enc.Encode(i.indexAbbrEncodableTridexes); err != nil {
		err = errors.Wrapf(err, "failed to write encoded object id")
		return
	}

	return
}

func (i *indexAbbr) readIfNecessary() (err error) {
	i.once.Do(
		func() {
			if i.didRead {
				return
			}

			ui.Log().Print("reading")

			i.didRead = true

			var r1 io.ReadCloser

			if r1, err = i.envRepo.ReadCloserCache(i.path); err != nil {
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

			if err = dec.Decode(&i.indexAbbrEncodableTridexes); err != nil {
				ui.Log().Print("finished decode unsuccessfully")
				err = errors.Wrap(err)
				return
			}
		},
	)

	return
}

func (i *indexAbbr) GetAbbr() (out ids.Abbr) {
	out.ZettelId.Expand = i.ZettelId().ExpandStringString
	out.BlobId.Expand = i.BlobId().ExpandStringString

	if i.AbbreviateZettelIds {
		out.ZettelId.Abbreviate = i.ZettelId().Abbreviate
	}

	if i.AbbreviateShas {
		out.BlobId.Abbreviate = i.BlobId().Abbreviate
	}

	return
}

func (i *indexAbbr) AddObjectToAbbreviationStore(
	o *sku.Transacted,
) (err error) {
	if err = i.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.hasChanges = true

	i.indexAbbrEncodableTridexes.BlobId.ObjectIds.Add(
		merkle.Format(o.GetBlobDigest()),
	)

	ks := o.GetObjectId().String()

	switch o.GetGenre() {
	case genres.Zettel:
		var h ids.ZettelId

		if err = h.SetFromIdParts(o.GetObjectId().Parts()); err != nil {
			err = errors.Wrap(err)
			return
		}

		i.indexAbbrEncodableTridexes.ZettelId.Heads.Add(h.GetHead())
		i.indexAbbrEncodableTridexes.ZettelId.Tails.Add(h.GetTail())

	case genres.Type,
		genres.Tag,
		genres.Config,
		genres.InventoryList,
		genres.Repo:
		return

	default:
		err = errors.ErrorWithStackf("unsupported object id: %#v", ks)
		return
	}

	return
}

func (i *indexAbbr) ZettelId() (asg sku.AbbrStoreGeneric[ids.ZettelId, *ids.ZettelId]) {
	asg = &i.indexAbbrEncodableTridexes.ZettelId

	return
}

func (i *indexAbbr) BlobId() (asg sku.AbbrStoreGeneric[merkle.Id, *merkle.Id]) {
	asg = &i.indexAbbrEncodableTridexes.BlobId

	return
}
