package zettel_id_index

import (
	"bufio"
	"encoding/gob"
	"io"
	"math/rand"
	"sync"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/coordinates"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/object_id_provider"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type index struct {
	su interfaces.CacheIOFactory

	lock *sync.RWMutex
	path string

	bitset collections.Bitset

	oldHinweisenStore *object_id_provider.Provider

	didRead    bool
	hasChanges bool

	nonRandomSelection bool
}

func MakeIndex(
	k interfaces.Config,
	s interfaces.Directory,
	su interfaces.CacheIOFactory,
) (i *index, err error) {
	i = &index{
		lock:               &sync.RWMutex{},
		path:               s.FileCacheObjectId(),
		nonRandomSelection: k.UsePredictableZettelIds(),
		su:                 su,
		bitset:             collections.MakeBitset(0),
	}

	if i.oldHinweisenStore, err = object_id_provider.New(s); err != nil {
		if errors.IsNotExist(err) {
			ui.TodoP4("determine which layer handles no-create kasten")
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (i *index) Flush() (err error) {
	i.lock.RLock()

	if !i.hasChanges {
		ui.Log().Print("no changes")
		i.lock.RUnlock()
		return
	}

	i.lock.RUnlock()

	var w1 io.WriteCloser

	if w1, err = i.su.WriteCloserCache(i.path); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.Deferred(&err, w1.Close)

	w := bufio.NewWriter(w1)

	defer errors.Deferred(&err, w.Flush)

	enc := gob.NewEncoder(w)

	if err = enc.Encode(i.bitset); err != nil {
		err = errors.Wrapf(err, "failed to write encoded zettel id")
		return
	}

	return
}

func (i *index) readIfNecessary() (err error) {
	i.lock.RLock()

	if i.didRead {
		i.lock.RUnlock()
		return
	}

	i.lock.RUnlock()

	i.lock.Lock()
	defer i.lock.Unlock()

	ui.Log().Print("reading")

	i.didRead = true

	var r1 io.ReadCloser

	if r1, err = i.su.ReadCloserCache(i.path); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	defer r1.Close()

	r := bufio.NewReader(r1)

	dec := gob.NewDecoder(r)

	if err = dec.Decode(i.bitset); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (i *index) Reset() (err error) {
	lMax := i.oldHinweisenStore.Left().Len() - 1
	rMax := i.oldHinweisenStore.Right().Len() - 1

	if lMax == 0 {
		err = errors.ErrorWithStackf("left zettel id are empty")
		return
	}

	if rMax == 0 {
		err = errors.ErrorWithStackf("right zettel id are empty")
		return
	}

	i.bitset = collections.MakeBitsetOn(lMax * rMax)

	i.hasChanges = true

	return
}

func (i *index) AddZettelId(k1 interfaces.ObjectId) (err error) {
	if !k1.GetGenre().EqualsGenre(genres.Zettel) {
		err = genres.MakeErrUnsupportedGenre(k1)
		return
	}

	var h ids.ZettelId

	if err = h.Set(k1.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = i.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var left, right int

	if left, err = i.oldHinweisenStore.Left().ZettelId(h.GetHead()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if right, err = i.oldHinweisenStore.Right().ZettelId(h.GetTail()); err != nil {
		err = errors.Wrap(err)
		return
	}

	k := coordinates.ZettelIdCoordinate{
		Left:  coordinates.Int(left),
		Right: coordinates.Int(right),
	}

	n := k.Id()
	ui.Log().Printf("deleting %d, %s", n, h)

	i.lock.Lock()
	defer i.lock.Unlock()

	i.bitset.DelIfPresent(int(n))

	i.hasChanges = true

	return
}

func (i *index) CreateZettelId() (h *ids.ZettelId, err error) {
	if err = i.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if i.bitset.CountOn() == 0 {
		err = errors.ErrorWithStackf("no available zettel ids")
		return
	}

	rand.Seed(time.Now().UnixNano())

	if i.bitset.CountOn() == 0 {
		err = errors.Wrap(object_id_provider.ErrZettelIdsExhausted{})
		return
	}

	ri := 0

	if i.bitset.CountOn() > 1 {
		ri = rand.Intn(i.bitset.CountOn() - 1)
	}

	m := 0
	j := 0

	if err = i.bitset.EachOff(
		func(n int) (err error) {
			if i.nonRandomSelection {
				if m == 0 {
					m = n
					return
				}

				if n > m {
					return
				}

				m = n
			} else {
				j++
				m = n

				if j == ri {
					err = collections.MakeErrStopIteration()
					return
				}
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	i.bitset.DelIfPresent(int(m))

	i.hasChanges = true

	return i.makeHinweisButDontStore(m)
}

func (i *index) makeHinweisButDontStore(
	j int,
) (h *ids.ZettelId, err error) {
	k := &coordinates.ZettelIdCoordinate{}
	k.SetInt(coordinates.Int(j))

	if h, err = ids.MakeZettelIdFromProvidersAndCoordinates(
		k.Id(),
		i.oldHinweisenStore.Left(),
		i.oldHinweisenStore.Right(),
	); err != nil {
		err = errors.Wrapf(err, "trying to make hinweis for %s, %d", k, j)
		return
	}

	return
}

func (i *index) PeekZettelIds(m int) (hs []*ids.ZettelId, err error) {
	if err = i.readIfNecessary(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if m > i.bitset.CountOn() || m == 0 {
		m = i.bitset.CountOn()
	}

	hs = make([]*ids.ZettelId, 0, m)
	j := 0

	if err = i.bitset.EachOff(
		func(n int) (err error) {
			n += 1
			k := &coordinates.ZettelIdCoordinate{}
			k.SetInt(coordinates.Int(n))

			var h *ids.ZettelId

			if h, err = i.makeHinweisButDontStore(n); err != nil {
				err = errors.Wrapf(err, "# %d", n)
				return
			}

			hs = append(hs, h)

			j++

			if j == m {
				err = collections.MakeErrStopIteration()
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
