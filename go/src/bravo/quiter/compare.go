package quiter

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type SortComparer[ELEMENT any] interface {
	SortCompare(ELEMENT, ELEMENT) SortCompare
}

type SortCompare interface {
	sortCompare()

	GetCompareInt() int
	Less() bool
	Equal() bool
	Greater() bool
}

type sortCompare int

func (sortCompare) sortCompare() {}

func (sortCompare sortCompare) GetCompareInt() int {
	return int(sortCompare)
}

func (sortCompare sortCompare) Less() bool {
	return sortCompare == SortCompareLess
}

func (sortCompare sortCompare) Equal() bool {
	return sortCompare == SortCompareEqual
}

func (sortCompare sortCompare) Greater() bool {
	return sortCompare == SortCompareGreater
}

const (
	SortCompareLess    = sortCompare(-1)
	SortCompareEqual   = sortCompare(0)
	SortCompareGreater = sortCompare(1)
)

func MakeSortComparerFromEqualerAndLessor3[ELEMENT any](
	equaler interfaces.Equaler[ELEMENT],
	lessor interfaces.Lessor3[ELEMENT],
) func(ELEMENT, ELEMENT) SortCompare {
	return func(left, right ELEMENT) SortCompare {
		if equaler.Equals(left, right) {
			return SortCompareEqual
		} else if lessor.Less(left, right) {
			return SortCompareLess
		} else {
			return SortCompareGreater
		}
	}
}
