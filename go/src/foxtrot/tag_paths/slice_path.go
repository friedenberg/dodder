package tag_paths

import (
	"slices"
)

type (
	SlicePaths []*Path
)

func (a *SlicePaths) Reset() {
	*a = (*a)[:0]
}

func (s SlicePaths) Len() int {
	return len(s)
}

func (s SlicePaths) Less(i, j int) bool {
	return s[i].Compare(s[j]) == -1
}

func (s SlicePaths) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s SlicePaths) ContainsPath(p *Path) (int, bool) {
	return slices.BinarySearchFunc(
		s,
		p,
		func(ep *Path, el *Path) int {
			return ep.Compare(p)
		},
	)
}

func (s *SlicePaths) AddNonEmptyPath(p *Path) {
	if p == nil || p.Len() == 1 {
		return
	}

	s.AddPath(p)
}

func (s *SlicePaths) AddPath(p *Path) (idx int, alreadyExists bool) {
	if p.IsEmpty() {
		return
	}

	// p = p.Clone()

	idx, alreadyExists = s.ContainsPath(p)

	if alreadyExists {
		return
	}

	*s = slices.Insert(*s, idx, p)

	return
}
