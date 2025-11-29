package ids

import (
	"slices"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
)

type (
	TagSet interface {
		Len() int
		All() interfaces.Seq[Tag]
		ContainsKey(string) bool
		Get(string) (Tag, bool)
		Key(Tag) string
	}

	TagSetMutable = interface {
		TagSet

		interfaces.Adder[Tag]
		DelKey(string) error
		interfaces.Resetable
	}
)

var TagSetEmpty TagSet

func init() {
	collections_ptr.RegisterGobValue[Tag](nil)
	TagSetEmpty = MakeTagSet()
}

func CloneTagSet(tags TagSet) TagSet {
	clone := MakeMutableTagSet()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
}

func CloneTagSetMutable(tags TagSet) TagSetMutable {
	clone := MakeMutableTagSet()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
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

func MakeMutableTagSet(hs ...Tag) TagSetMutable {
	return MakeTagMutableSet(hs...)
}

func MakeTagMutableSet(hs ...Tag) TagSetMutable {
	return TagSetMutable(
		collections_ptr.MakeMutableValueSetValue(
			nil,
			hs...,
		),
	)
}

func TagSetEquals(first, second TagSet) bool {
	return quiter.SetEquals(first, second)
}

type TagSlice []Tag

func MakeTagSlice(tags ...Tag) (slice TagSlice) {
	slice = make([]Tag, len(tags))

	copy(slice, tags)

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

func (slice *TagSlice) DropFirst() {
	if slice.Len() > 0 {
		*slice = (*slice)[1:]
	}
}

func (slice TagSlice) Len() int {
	return len(slice)
}

func (slice *TagSlice) AddString(value string) (err error) {
	var tag Tag

	if err = tag.Set(value); err != nil {
		err = errors.Wrap(err)
		return err
	}

	slice.Add(tag)

	return err
}

func (slice *TagSlice) Add(e Tag) {
	*slice = append(*slice, e)
}

func (slice *TagSlice) Set(v string) (err error) {
	tags := strings.SplitSeq(v, ",")

	for tag := range tags {
		if err = slice.AddString(tag); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (slice TagSlice) SortedString() (out []string) {
	out = make([]string, len(slice))

	i := 0

	for _, e := range slice {
		out[i] = e.String()
		i++
	}

	slices.Sort(out)

	return out
}

func (slice TagSlice) String() string {
	return strings.Join(slice.SortedString(), ", ")
}

func (slice TagSlice) ToSet() TagSet {
	return MakeTagSet([]Tag(slice)...)
}
