package collections

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO-P3 move to iter
func MakeWriterNoop[T any]() interfaces.FuncIter[T] {
	return func(e T) (err error) {
		return
	}
}

// TODO-P3 move to iter
func MakeTryFinally[T any](
	try interfaces.FuncIter[T],
	finally interfaces.FuncIter[T],
) interfaces.FuncIter[T] {
	return func(e T) (err error) {
		defer func() {
			err1 := finally(e)

			if err != nil {
				err = errors.MakeMulti(err, err1)
			} else {
				err = err1
			}
		}()

		if err = try(e); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}
