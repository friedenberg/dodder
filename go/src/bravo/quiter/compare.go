package quiter

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
