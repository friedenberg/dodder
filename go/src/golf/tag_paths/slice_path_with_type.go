package tag_paths

import (
	"slices"
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

func (pathsWithTypes PathsWithTypes) Less(i, j int) bool {
	return pathsWithTypes[i].Compare(&pathsWithTypes[j].Path) == -1
}

func (pathsWithTypes PathsWithTypes) Swap(i, j int) {
	pathsWithTypes[j], pathsWithTypes[i] = pathsWithTypes[i], pathsWithTypes[j]
}

func (pathsWithTypes PathsWithTypes) ContainsPath(p *PathWithType) (int, bool) {
	return slices.BinarySearchFunc(
		pathsWithTypes,
		p,
		func(ep *PathWithType, el *PathWithType) int {
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
