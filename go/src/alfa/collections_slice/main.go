package collections_slice

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type Slice[ELEMENT any] []ELEMENT

func (slice Slice[ELEMENT]) Len() int {
	return len(slice)
}

func (slice Slice[ELEMENT]) Any() (element ELEMENT) {
	if slice.Len() > 0 {
		element = slice[0]
	}

	return element
}

func (slice Slice[ELEMENT]) All() interfaces.Seq[ELEMENT] {
	return func(yield func(ELEMENT) bool) {
		for _, element := range slice {
			if !yield(element) {
				break
			}
		}
	}
}

func (slice *Slice[ELEMENT]) Grow(n int) {
	*slice = slices.Grow(*slice, n)
}

func (slice *Slice[ELEMENT]) Append(elements ...ELEMENT) {
	*slice = append(*slice, elements...)
}

func (slice *Slice[ELEMENT]) Reset() {
	*slice = (*slice)[:0]
}

func (slice *Slice[ELEMENT]) ResetWith(other interfaces.Seq[ELEMENT]) {
	slice.Reset()

	if otherSlice, ok := any(other).(Slice[ELEMENT]); ok {
		slice.Grow(otherSlice.Len())

		for _, element := range otherSlice {
			slice.Append(element)
		}
	} else {
		for element := range other {
			slice.Append(element)
		}
	}
}
