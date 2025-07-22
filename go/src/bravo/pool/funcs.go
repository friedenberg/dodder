package pool

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO rewrite with iterator?
func MakePooledChain[T any, TPtr interfaces.Ptr[T]](
	pool interfaces.Pool[T, TPtr],
	funcIters ...interfaces.FuncIter[TPtr],
) interfaces.FuncIter[TPtr] {
	return func(element TPtr) (err error) {
		for _, w := range funcIters {
			err = w(element)

			switch {
			case err == nil:
				continue

			case IsDoNotRepool(err):
				err = nil
				return

			case errors.IsStopIteration(err):
				err = nil
				pool.Put(element)
				return

			default:
				pool.Put(element)
				err = errors.Wrap(err)
				return
			}
		}

		pool.Put(element)

		return
	}
}
