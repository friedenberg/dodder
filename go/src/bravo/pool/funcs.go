package pool

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MakeWriterDoNotRepool[T any]() interfaces.FuncIter[*T] {
	return func(e *T) (err error) {
		err = ErrDoNotRepool
		return
	}
}

func MakePooledChain[T interfaces.Poolable[T], TPtr interfaces.PoolablePtr[T]](
	p interfaces.Pool[T, TPtr],
	wfs ...interfaces.FuncIter[TPtr],
) interfaces.FuncIter[TPtr] {
	return func(e TPtr) (err error) {
		for _, w := range wfs {
			err = w(e)

			switch {
			case err == nil:
				continue

			case IsDoNotRepool(err):
				err = nil
				return

			case errors.IsStopIteration(err):
				err = nil
				p.Put(e)
				return

			default:
				p.Put(e)
				err = errors.Wrap(err)
				return
			}
		}

		p.Put(e)

		return
	}
}
