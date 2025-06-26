package pool

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type poolWithReset[T any, TPtr interfaces.Resetable[T]] struct {
	inner *sync.Pool
}

func MakePoolWithReset[T any, TPtr interfaces.Resetable[T]]() *poolWithReset[T, TPtr] {
	return &poolWithReset[T, TPtr]{
		inner: &sync.Pool{
			New: func() interface{} {
				o := new(T)
				TPtr(o).Reset()

				return o
			},
		},
	}
}

func (p poolWithReset[T, TPtr]) Apply(f interfaces.FuncIter[T], e T) (err error) {
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

func (ip poolWithReset[T, TPtr]) Get() TPtr {
	return ip.inner.Get().(TPtr)
}

func (ip poolWithReset[T, TPtr]) PutMany(is ...TPtr) (err error) {
	for _, i := range is {
		if err = ip.Put(i); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (ip poolWithReset[T, TPtr]) Put(i TPtr) (err error) {
	if i == nil {
		return
	}

	i.Reset()
	ip.inner.Put(i)

	return
}
