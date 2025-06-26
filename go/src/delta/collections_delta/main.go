package collections_delta

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
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

func MakeSetDelta[T interfaces.ValueLike](
	from, to interfaces.SetLike[T],
) interfaces.Delta[T] {
	d := delta[T]{
		Added:   collections_value.MakeMutableValueSet[T](nil),
		Removed: from.CloneMutableSetLike(),
	}

	to.Each(
		func(e T) (err error) {
			if from.Contains(e) {
				// had previously
			} else {
				// did not have previously
				d.Added.Add(e)
			}

			d.Removed.Del(e)

			return
		},
	)

	return d
}
