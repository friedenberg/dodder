package pool

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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
				return err

			case errors.IsStopIteration(err):
				err = nil
				pool.Put(element)
				return err

			default:
				pool.Put(element)
				err = errors.Wrap(err)
				return err
			}
		}

		pool.Put(element)

		return err
	}
}
