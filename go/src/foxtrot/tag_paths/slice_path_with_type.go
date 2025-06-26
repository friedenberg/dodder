package tag_paths

import (
	"slices"
)

type (
	PathsWithTypes []*PathWithType
)

func (a *PathsWithTypes) Reset() {
	*a = (*a)[:0]
}

func (s PathsWithTypes) Len() int {
	return len(s)
}

func (s PathsWithTypes) Less(i, j int) bool {
	return s[i].Compare(&s[j].Path) == -1
}

func (s PathsWithTypes) Swap(i, j int) {
	s[j], s[i] = s[i], s[j]
}

func (s PathsWithTypes) ContainsPath(p *PathWithType) (int, bool) {
	return slices.BinarySearchFunc(
		s,
		p,
		func(ep *PathWithType, el *PathWithType) int {
			return ep.Compare(&p.Path)
		},
	)
}

func (s *PathsWithTypes) AddNonEmptyPath(p *PathWithType) {
	if p == nil {
		return
	}

	s.AddPath(p)
}

func (s *PathsWithTypes) AddPath(p *PathWithType) (idx int, alreadyExists bool) {
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
