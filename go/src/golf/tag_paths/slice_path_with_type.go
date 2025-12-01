package tag_paths

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/cmp"
)

type (
	PathsWithTypes []*PathWithType
)

func (pathsWithTypes *PathsWithTypes) Reset() {
	*pathsWithTypes = (*pathsWithTypes)[:0]
}

func (pathsWithTypes PathsWithTypes) Len() int {
	return len(pathsWithTypes)
}

func (pathsWithTypes PathsWithTypes) Less(left, right int) bool {
	return pathsWithTypes[left].Compare(&pathsWithTypes[right].Path).IsLess()
}

func (pathsWithTypes PathsWithTypes) Swap(left, right int) {
	pathsWithTypes[right], pathsWithTypes[left] = pathsWithTypes[left], pathsWithTypes[right]
}

func (pathsWithTypes PathsWithTypes) ContainsPath(p *PathWithType) (int, bool) {
	return cmp.BinarySearchFunc(
		pathsWithTypes,
		p,
		func(ep *PathWithType, el *PathWithType) cmp.Result {
			return ep.Compare(&p.Path)
		},
	)
}

func (pathsWithTypes *PathsWithTypes) AddNonEmptyPath(p *PathWithType) {
	if p == nil {
		return
	}

	pathsWithTypes.AddPath(p)
}

func (pathsWithTypes *PathsWithTypes) AddPath(p *PathWithType) (idx int, alreadyExists bool) {
	if p.IsEmpty() {
		return idx, alreadyExists
	}

	// p = p.Clone()

	idx, alreadyExists = pathsWithTypes.ContainsPath(p)

	if alreadyExists {
		return idx, alreadyExists
	}

	*pathsWithTypes = slices.Insert(*pathsWithTypes, idx, p)

	return idx, alreadyExists
}
