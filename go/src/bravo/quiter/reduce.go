package quiter

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Elements[T any](s interfaces.Collection[T]) (out []T) {
	if s == nil {
		return out
	}

	out = make([]T, 0, s.Len())

	for v := range s.All() {
		out = append(out, v)
	}

	return out
}

func ElementsSorted[T any](
	s interfaces.Collection[T],
	sf func(T, T) bool,
) (out []T) {
	if s == nil {
		return out
	}

	out = make([]T, 0, s.Len())

	for v := range s.All() {
		out = append(out, v)
	}

	sort.Slice(out, func(i, j int) bool {
		return sf(out[i], out[j])
	})

	return out
}
