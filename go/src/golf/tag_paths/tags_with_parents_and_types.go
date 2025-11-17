package tag_paths

import (
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type (
	TagsWithParentsAndTypes []TagWithParentsAndTypes
)

func (s *TagsWithParentsAndTypes) Reset() {
	*s = (*s)[:0]
}

func (s TagsWithParentsAndTypes) Len() int {
	return len(s)
}

func (s TagsWithParentsAndTypes) ContainsObjectIdTag(
	k *ids.ObjectId,
) (int, bool) {
	return s.containsObjectIdTag(k, true)
}

func (s TagsWithParentsAndTypes) ContainsObjectIdTagExact(
	k *ids.ObjectId,
) (int, bool) {
	return s.containsObjectIdTag(k, false)
}

// TODO make less fragile
func (s TagsWithParentsAndTypes) containsObjectIdTag(
	k *ids.ObjectId,
	partial bool,
) (int, bool) {
	e := k.PartsStrings().Right
	offset := 0

	if k.IsVirtual() {
		percent := catgut.GetPool().Get()
		defer catgut.GetPool().Put(percent)

		percent.Set("%")

		loc, ok := slices.BinarySearchFunc(
			s,
			percent,
			func(ewp TagWithParentsAndTypes, e *Tag) int {
				cmp := ewp.Tag.ComparePartial(e)
				return cmp
			},
		)

		if !ok {
			return loc, ok
		}

		offset = percent.Len()
		s = s[loc:]
	}

	return slices.BinarySearchFunc(
		s,
		e,
		func(ewp TagWithParentsAndTypes, e *Tag) int {
			cmp := catgut.CompareUTF8Bytes(
				catgut.ComparerBytes(ewp.Tag.Bytes()[offset:]),
				catgut.ComparerBytes(e.Bytes()),
				partial,
			)

			return cmp
		},
	)
}

func (s TagsWithParentsAndTypes) ContainsTag(e *Tag) (int, bool) {
	return slices.BinarySearchFunc(
		s,
		e,
		func(ewp TagWithParentsAndTypes, e *Tag) int {
			cmp := ewp.Tag.ComparePartial(e)
			return cmp
		},
	)
}

func (s TagsWithParentsAndTypes) ContainsString(
	value string,
) (int, bool) {
	return slices.BinarySearchFunc(
		s,
		catgut.ComparerString(value),
		func(ewp TagWithParentsAndTypes, c catgut.ComparerString) int {
			cmp := catgut.CompareUTF8Bytes(
				catgut.ComparerBytes(ewp.Tag.Bytes()),
				c,
				true,
			)
			return cmp
		},
	)
}

func (s TagsWithParentsAndTypes) GetMatching(
	e *Tag,
) (matching []TagWithParentsAndTypes) {
	i, ok := s.ContainsTag(e)

	if !ok {
		return matching
	}

	for _, ewp := range s[i:] {
		cmp := ewp.ComparePartial(e)

		if cmp != 0 {
			return matching
		}

		matching = append(matching, ewp)
	}

	return matching
}

// TODO return success
func (s *TagsWithParentsAndTypes) Add(
	e1 *Tag,
	p *PathWithType,
) (err error) {
	var e *Tag

	if e, err = e1.Clone(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	idx, ok := s.ContainsTag(e)

	var a TagWithParentsAndTypes

	if ok {
		a = (*s)[idx]
		a.Parents.AddNonEmptyPath(p)
		(*s)[idx] = a
	} else {
		a = TagWithParentsAndTypes{Tag: e}
		a.Parents.AddNonEmptyPath(p)

		if idx == s.Len() {
			*s = append(*s, a)
		} else {
			*s = slices.Insert(*s, idx, a)
		}
	}

	return err
}

// TODO return success
func (s *TagsWithParentsAndTypes) Remove(e1 *Tag) (err error) {
	var e *Tag

	if e, err = e1.Clone(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	idx, ok := s.ContainsTag(e)

	if !ok {
		return err
	}

	*s = slices.Delete(*s, idx, idx+1)

	return err
}

func (s TagsWithParentsAndTypes) StringCommaSeparatedExplicit() string {
	var sb strings.Builder

	first := true

	for _, ewp := range s {
		if ewp.Parents.Len() != 0 {
			continue
		}

		sb.Write(ewp.Tag.Bytes())

		if !first {
			sb.WriteString(", ")
		}

		first = false
	}

	return sb.String()
}
