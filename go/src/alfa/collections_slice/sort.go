package collections_slice

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
)

func (slice *Slice[ELEMENT]) SortByStringFunc(getKey func(ELEMENT) string) {
	sort.Slice(
		*slice,
		func(i, j int) bool {
			return getKey(slice.At(i)) < getKey(slice.At(j))
		},
	)
}

func (slice *Slice[ELEMENT]) SortWithComparer(cmp cmp.Func[ELEMENT]) {
	sort.Slice(
		*slice,
		func(i, j int) bool {
			return cmp(slice.At(i), slice.At(j)).IsLess()
		},
	)
}
