package quiter

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
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

func SetEqualsPtr[ELEMENT any, ELEMENT_PTR interfaces.Ptr[ELEMENT]](
	a, b interfaces.SetPtrLike[ELEMENT, ELEMENT_PTR],
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

	for element := range a.All() {
		key := b.Key(element)

		if ok := b.ContainsKey(key); !ok {
			return false
		}
	}

	return true
}
