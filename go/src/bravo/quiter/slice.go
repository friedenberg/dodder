package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO move to own package
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

func (slice *Slice[ELEMENT]) Append(elements ...ELEMENT) {
	*slice = append(*slice, elements...)
}
