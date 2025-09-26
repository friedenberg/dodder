package zettel_id_index

import (
	"bufio"
	"encoding/gob"
	"io"
	"math/rand"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/coordinates"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

type encodedIds struct {
	AvailableIds map[int]bool
}

type index struct {
	namedBlobAccess interfaces.NamedBlobAccess

	lock sync.Mutex
	path string

	encodedIds

	oldZettelIdStore *object_id_provider.Provider

	didRead    bool
	hasChanges bool

	nonRandomSelection bool
}

func MakeIndex(
	configCli repo_config_cli.Config,
	dir interfaces.Directory,
	namedBlobAccess interfaces.NamedBlobAccess,
) (i *index, err error) {
	i = &index{
		path:               dir.FileCacheObjectId(),
		nonRandomSelection: configCli.UsePredictableZettelIds(),
		namedBlobAccess:    namedBlobAccess,
		encodedIds: encodedIds{
			AvailableIds: make(map[int]bool),
		},
	}

	if i.oldZettelIdStore, err = object_id_provider.New(dir); err != nil {
		if errors.IsNotExist(err) {
			ui.TodoP4("determine which layer handles no-create kasten")
			err = nil
		} else {
			err = errors.Wrap(err)
			return i, err
		}
	}

	return i, err
}

func (index *index) Flush() (err error) {
	index.lock.Lock()
	defer index.lock.Unlock()

	if !index.hasChanges {
		ui.Log().Print("no changes")
		return err
	}

	var namedBlobWriter interfaces.BlobWriter

	if namedBlobWriter, err = index.namedBlobAccess.MakeNamedBlobWriter(
		index.path,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, namedBlobWriter)

	w := bufio.NewWriter(namedBlobWriter)

	defer errors.DeferredFlusher(&err, w)

	enc := gob.NewEncoder(w)

	if err = enc.Encode(index.encodedIds); err != nil {
		err = errors.Wrapf(err, "failed to write encoded object id")
		return err
	}

	return err
}

func (index *index) readIfNecessary() (err error) {
	index.lock.Lock()
	defer index.lock.Unlock()

	if index.didRead {
		return err
	}

	ui.Log().Print("reading")

	index.didRead = true

	var namedBlobReader interfaces.BlobReader

	if namedBlobReader, err = index.namedBlobAccess.MakeNamedBlobReader(
		index.path,
	); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	defer namedBlobReader.Close()

	r := bufio.NewReader(namedBlobReader)

	dec := gob.NewDecoder(r)

	if err = dec.Decode(&index.encodedIds); err != nil {
		err = errors.WrapExceptSentinelAsNil(err, io.EOF)
		return err
	}

	return err
}

func (index *index) Reset() (err error) {
	index.lock.Lock()
	defer index.lock.Unlock()

	lMax := index.oldZettelIdStore.Left().Len() - 1
	rMax := index.oldZettelIdStore.Right().Len() - 1

	if lMax == 0 {
		err = errors.ErrorWithStackf("left object id are empty")
		return err
	}

	if rMax == 0 {
		err = errors.ErrorWithStackf("right object id are empty")
		return err
	}

	index.AvailableIds = make(map[int]bool, lMax*rMax)

	for l := 0; l <= lMax; l++ {
		for r := 0; r <= rMax; r++ {
			k := &coordinates.ZettelIdCoordinate{
				Left:  coordinates.Int(l),
				Right: coordinates.Int(r),
			}

			ui.Log().Print(k)

			n := int(k.Id())
			index.AvailableIds[n] = true
		}
	}

	index.hasChanges = true

	return err
}

func (index *index) AddZettelId(k1 interfaces.ObjectId) (err error) {
	if !k1.GetGenre().EqualsGenre(genres.Zettel) {
		err = genres.MakeErrUnsupportedGenre(k1)
		return err
	}

	var h ids.ZettelId

	if err = h.SetFromIdParts(k1.Parts()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = index.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var left, right int

	if left, err = index.oldZettelIdStore.Left().ZettelId(h.GetHead()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if right, err = index.oldZettelIdStore.Right().ZettelId(h.GetTail()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	k := coordinates.ZettelIdCoordinate{
		Left:  coordinates.Int(left),
		Right: coordinates.Int(right),
	}

	n := k.Id()
	ui.Log().Printf("deleting %d, %s", n, h)

	index.lock.Lock()
	defer index.lock.Unlock()

	delete(index.AvailableIds, int(n))

	index.hasChanges = true

	return err
}

func (index *index) CreateZettelId() (h *ids.ZettelId, err error) {
	if err = index.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return h, err
	}

	if len(index.AvailableIds) == 0 {
		err = errors.Wrap(object_id_provider.ErrZettelIdsExhausted{})
		return h, err
	}

	ri := 0

	if len(index.AvailableIds) > 1 {
		ri = rand.Intn(len(index.AvailableIds) - 1)
	}

	m := 0
	j := 0

	for n := range index.AvailableIds {
		if index.nonRandomSelection {
			if m == 0 {
				m = n
				continue
			}

			if n > m {
				continue
			}

			m = n
		} else {
			j++
			m = n

			if j == ri {
				break
			}
		}
	}

	delete(index.AvailableIds, int(m))

	index.hasChanges = true

	return index.makeZettelIdButDontStore(m)
}

func (index *index) makeZettelIdButDontStore(
	j int,
) (h *ids.ZettelId, err error) {
	k := &coordinates.ZettelIdCoordinate{}
	k.SetInt(coordinates.Int(j))

	h, err = ids.MakeZettelIdFromProvidersAndCoordinates(
		k.Id(),
		index.oldZettelIdStore.Left(),
		index.oldZettelIdStore.Right(),
	)
	if err != nil {
		err = errors.Wrapf(err, "trying to make hinweis for %s", k)
		return h, err
	}

	return h, err
}

func (index *index) PeekZettelIds(m int) (hs []*ids.ZettelId, err error) {
	if err = index.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return hs, err
	}

	if m > len(index.AvailableIds) || m == 0 {
		m = len(index.AvailableIds)
	}

	hs = make([]*ids.ZettelId, 0, m)
	j := 0

	for n := range index.AvailableIds {
		k := &coordinates.ZettelIdCoordinate{}
		k.SetInt(coordinates.Int(n))

		var h *ids.ZettelId

		if h, err = index.makeZettelIdButDontStore(n); err != nil {
			err = errors.Wrap(err)
			return hs, err
		}

		hs = append(hs, h)

		j++

		if j == m {
			break
		}
	}

	return hs, err
}
