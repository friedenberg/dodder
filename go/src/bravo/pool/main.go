package pool

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type pool[T any, TPtr interfaces.Ptr[T]] struct {
	inner *sync.Pool
	reset func(TPtr)
}

func MakePool[T any, TPtr interfaces.Ptr[T]](
	New func() TPtr,
	Reset func(TPtr),
) *pool[T, TPtr] {
	return &pool[T, TPtr]{
		reset: Reset,
		inner: &sync.Pool{
			New: func() (swimmer any) {
				if New == nil {
					swimmer = new(T)
				} else {
					swimmer = New()
				}

				return
			},
		},
	}
}

func (pool pool[T, TPtr]) Apply(funk interfaces.FuncIter[T], e T) (err error) {
	err = funk(e)

	switch {

	case IsDoNotRepool(err):
		err = nil
		return

	case errors.IsStopIteration(err):
		err = nil
		pool.Put(&e)

	case err != nil:
		err = errors.Wrap(err)

		fallthrough

	default:
		pool.Put(&e)
	}

	return
}

func (pool pool[T, TPtr]) Get() TPtr {
	return pool.inner.Get().(TPtr)
}

func (pool pool[T, TPtr]) PutMany(swimmers ...TPtr) (err error) {
	for _, i := range swimmers {
		if err = pool.Put(i); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (pool pool[T, TPtr]) Put(swimmer TPtr) (err error) {
	if swimmer == nil {
		return
	}

	if pool.reset != nil {
		pool.reset(swimmer)
	}

	pool.inner.Put(swimmer)

	return
}
