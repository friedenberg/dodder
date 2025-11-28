package collections_delta

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

type delta[T interfaces.ValueLike] struct {
	Added, Removed interfaces.MutableSetLike[T]
}

func (d delta[T]) GetAdded() interfaces.SetLike[T] {
	return d.Added
}

func (d delta[T]) GetRemoved() interfaces.SetLike[T] {
	return d.Removed
}

func MakeSetDelta[ELEMENT interfaces.ValueLike](
	from, to interfaces.SetBase[ELEMENT],
) interfaces.Delta[ELEMENT] {
	delta := delta[ELEMENT]{
		Added:   collections_value.MakeMutableValueSet[ELEMENT](nil),
		Removed: collections_value.MakeMutableValueSet[ELEMENT](nil),
	}

	for element := range from.All() {
		delta.Removed.Add(element)
	}

	for element := range to.All() {
		if from.Contains(element) {
			// had previously
		} else {
			// did not have previously
			delta.Added.Add(element)
		}

		delta.Removed.Del(element)
	}

	return delta
}
