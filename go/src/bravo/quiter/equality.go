package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO refactor to use iterators
func SetEquals[T any](
	a, b interfaces.SetLike[T],
) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil {
		return b.Len() == 0
	}

	if b == nil {
		return a.Len() == 0
	}

	if a.Len() != b.Len() {
		return false
	}

	err := a.Each(
		func(e T) (err error) {
			if ok := b.Contains(e); !ok {
				err = errors.ErrFalse
				return
			}

			return
		},
	)
	if err != nil {
		return false
	}

	return true
}

func SetEqualsPtr[T any, TPtr interfaces.Ptr[T]](
	a, b interfaces.SetPtrLike[T, TPtr],
) bool {
	if a == nil && b == nil {
		return true
	}

	if a == nil {
		return false
	}

	if a.Len() != b.Len() {
		return false
	}

	err := a.EachPtr(
		func(e TPtr) (err error) {
			k := b.KeyPtr(e)

			if ok := b.ContainsKey(k); !ok {
				err = errors.ErrFalse
				return
			}

			return
		},
	)
	if err != nil {
		return false
	}

	return true
}
