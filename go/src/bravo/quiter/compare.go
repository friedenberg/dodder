package quiter

type SortCompare interface {
	Less() bool
	Equal() bool
	Greater() bool
	sortCompare()
}

type sortCompare int

func (sortCompare) sortCompare() {}

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
	SortCompareLess = sortCompare(iota)
	SortCompareEqual
	SortCompareGreater
)
