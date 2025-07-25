package quiter

import (
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

	for e := range a.All() {
		if ok := b.Contains(e); !ok {
			return false
		}
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

	for e := range a.AllPtr() {
		k := b.KeyPtr(e)

		if ok := b.ContainsKey(k); !ok {
			return false
		}
	}

	return true
}
