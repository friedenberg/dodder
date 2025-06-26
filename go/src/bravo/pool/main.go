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
			New: func() (o interface{}) {
				if New == nil {
					o = new(T)
				} else {
					o = New()
				}

				return
			},
		},
	}
}

func (p pool[T, TPtr]) Apply(f interfaces.FuncIter[T], e T) (err error) {
	err = f(e)

	switch {

	case IsDoNotRepool(err):
		err = nil
		return

	case errors.IsStopIteration(err):
		err = nil
		p.Put(&e)

	case err != nil:
		err = errors.Wrap(err)

		fallthrough

	default:
		p.Put(&e)
	}

	return
}

func (ip pool[T, TPtr]) Get() TPtr {
	return ip.inner.Get().(TPtr)
}

func (ip pool[T, TPtr]) PutMany(is ...TPtr) (err error) {
	for _, i := range is {
		if err = ip.Put(i); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (ip pool[T, TPtr]) Put(i TPtr) (err error) {
	if i == nil {
		return
	}

	if ip.reset != nil {
		ip.reset(i)
	}

	ip.inner.Put(i)

	return
}
