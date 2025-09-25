package ids

import (
	"sort"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
)

type (
	TagSet        = interfaces.SetPtrLike[Tag, *Tag]
	TagMutableSet = interfaces.MutableSetPtrLike[Tag, *Tag]
)

var TagSetEmpty TagSet

func init() {
	collections_ptr.RegisterGobValue[Tag](nil)
	TagSetEmpty = MakeTagSet()
}

func MakeTagSet(es ...Tag) (s TagSet) {
	if len(es) == 0 && TagSetEmpty != nil {
		return TagSetEmpty
	}

	return TagSet(
		collections_ptr.MakeValueSetValue(nil, es...),
	)
}

func MakeTagSetStrings(vs ...string) (s TagSet, err error) {
	return collections_ptr.MakeValueSetString[Tag](nil, vs...)
}

func MakeMutableTagSet(hs ...Tag) TagMutableSet {
	return MakeTagMutableSet(hs...)
}

func MakeTagMutableSet(hs ...Tag) TagMutableSet {
	return TagMutableSet(
		collections_ptr.MakeMutableValueSetValue(
			nil,
			hs...,
		),
	)
}

func TagSetEquals(first, second TagSet) bool {
	return quiter.SetEqualsPtr(first, second)
}

type TagSlice []Tag

func MakeTagSlice(tags ...Tag) (slice TagSlice) {
	slice = make([]Tag, len(tags))

	for index, tag := range tags {
		slice[index] = tag
	}

	return slice
}

func NewSliceFromStrings(tagStrings ...string) (slice TagSlice, err error) {
	slice = make([]Tag, len(tagStrings))

	for index, tagString := range tagStrings {
		if err = slice[index].Set(tagString); err != nil {
			err = errors.Wrap(err)
			return slice, err
		}
	}

	return slice, err
}

func (s *TagSlice) DropFirst() {
	if s.Len() > 0 {
		*s = (*s)[1:]
	}
}

func (s TagSlice) Len() int {
	return len(s)
}

func (tags *TagSlice) AddString(value string) (err error) {
	var tag Tag

	if err = tag.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	tags.Add(tag)

	return err
}

func (es *TagSlice) Add(e Tag) {
	*es = append(*es, e)
}

func (s *TagSlice) Set(v string) (err error) {
	es := strings.Split(v, ",")

	for _, e := range es {
		if err = s.AddString(e); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (es TagSlice) SortedString() (out []string) {
	out = make([]string, len(es))

	i := 0

	for _, e := range es {
		out[i] = e.String()
		i++
	}

	sort.Slice(
		out,
		func(i, j int) bool {
			return out[i] < out[j]
		},
	)

	return out
}

func (s TagSlice) String() string {
	return strings.Join(s.SortedString(), ", ")
}

func (s TagSlice) ToSet() TagSet {
	return MakeTagSet([]Tag(s)...)
}
