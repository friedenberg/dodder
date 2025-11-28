package quiter

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type SetBaseContainer[ELEMENT interfaces.Stringer] struct {
	interfaces.SetBase[ELEMENT]
}

func (set SetBaseContainer[ELEMENT]) Contains(element ELEMENT) bool {
	return set.ContainsKey(element.String())
}

// TODO refactor to use iterators
func SetEquals[ELEMENT interfaces.Stringer](
	a, b interfaces.SetBase[ELEMENT],
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
		if ok := b.ContainsKey(e.String()); !ok {
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
