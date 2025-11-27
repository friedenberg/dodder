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

func (slice *Slice[ELEMENT]) Merge(otherSeq Slice[ELEMENT]) {
	*slice = append(*slice, otherSeq...)
}

func (slice *Slice[ELEMENT]) Reset() {
	*slice = (*slice)[:0]
}

func (slice *Slice[ELEMENT]) ResetWith(other Slice[ELEMENT]) {
	slice.ResetWithCollection(other)
}

func (slice *Slice[ELEMENT]) ResetWithCollection(
	collection interfaces.Collection[ELEMENT],
) {
	slice.Reset()
	slice.Grow(collection.Len())

	for element := range collection.All() {
		slice.Append(element)
	}
}

func (slice *Slice[ELEMENT]) ResetWithSeq(other interfaces.Seq[ELEMENT]) {
	slice.Reset()

	for element := range other {
		slice.Append(element)
	}
}
