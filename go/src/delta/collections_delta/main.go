package collections_delta

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
)

type delta[T interfaces.ValueLike] struct {
	Added, Removed interfaces.MutableSetLike[T]
}

func (d delta[T]) GetAdded() interfaces.Set[T] {
	return d.Added
}

func (d delta[T]) GetRemoved() interfaces.Set[T] {
	return d.Removed
}

func MakeSetDelta[ELEMENT interfaces.ValueLike](
	from, to interfaces.Set[ELEMENT],
) interfaces.Delta[ELEMENT] {
	delta := delta[ELEMENT]{
		Added:   collections_value.MakeMutableValueSet[ELEMENT](nil),
		Removed: collections_value.MakeMutableValueSet[ELEMENT](nil),
	}

	for element := range from.All() {
		delta.Removed.Add(element)
	}

	for element := range to.All() {
		if from.ContainsKey(element.String()) {
			// had previously
		} else {
			// did not have previously
			delta.Added.Add(element)
		}

		quiter_set.Del(delta.Removed, element)
	}

	return delta
}
