package ids

import (
	_ "encoding/gob"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
)

type (
	TagSlice      = collections_slice.Slice[TagStruct]
	TagSet        = Set[TagStruct]
	TagSetMutable = SetMutable[TagStruct]
)

var TagSetEmpty TagSet

func init() {
	collections_ptr.RegisterGobValue[TagStruct](nil)
	TagSetEmpty = collections_ptr.MakeValueSetValue[TagStruct](nil)
}

// TODO move to quiter
func CloneTagSet(tags TagSet) TagSet {
	clone := MakeTagSetMutable()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
}

// TODO move to quiter
func CloneTagSetMutable(tags TagSet) TagSetMutable {
	clone := MakeTagSetMutable()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
}

// TODO move to quiter
func MakeTagSetFromSlice(tags ...TagStruct) (s TagSet) {
	if len(tags) == 0 {
		return TagSetEmpty
	}

	return collections_ptr.MakeValueSetValue(nil, tags...)
}

// TODO move to quiter
func MakeTagSetStrings(tagStrings ...string) (s TagSet, err error) {
	return collections_ptr.MakeValueSetString[TagStruct](nil, tagStrings...)
}

// TODO move to quiter
func MakeTagSetMutable(tags ...TagStruct) TagSetMutable {
	return collections_ptr.MakeMutableValueSetValue(nil, tags...)
}

func ExpandTagSet(set TagSet, expander expansion.Expander) TagSetMutable {
	setMutable := MakeTagSetMutable()

	for tag := range expansion.ExpandMany(set.All(), expander) {
		setMutable.Add(tag)
	}

	return setMutable
}

func IntersectPrefixes(haystack TagSet, needle TagStruct) (s3 TagSet) {
	s4 := MakeTagSetMutable()

	for _, e := range quiter.CollectSlice[TagStruct](haystack) {
		if strings.HasPrefix(e.String(), needle.String()) {
			s4.Add(e)
		}
	}

	s3 = CloneTagSet(s4)

	return s3
}

func SubtractPrefix(s1 TagSet, e TagStruct) (s2 TagSet) {
	s3 := MakeTagSetMutable()

	for _, e1 := range quiter.CollectSlice[TagStruct](s1) {
		e2, _ := LeftSubtract(e1, e)

		if e2.String() == "" {
			continue
		}

		s3.Add(e2)
	}

	s2 = CloneTagSet(s3)

	return s2
}

func WithRemovedCommonPrefixes(tags TagSet) (output TagSet) {
	sortedTags := quiter.SortedValues(tags.All())
	filteredTags := make([]TagStruct, 0, len(sortedTags))

	for _, e := range sortedTags {
		if len(filteredTags) == 0 {
			filteredTags = append(filteredTags, e)
			continue
		}

		idxLast := len(filteredTags) - 1
		last := filteredTags[idxLast]

		switch {
		case Contains(last, e):
			continue

		case Contains(e, last):
			filteredTags[idxLast] = e

		default:
			filteredTags = append(filteredTags, e)
		}
	}

	output = MakeTagSetFromSlice(filteredTags...)

	return output
}

func AddNormalizedTag(tags TagSetMutable, tag TagStruct) {
	seq := expansion.ExpandOneIntoIds[TagStruct](
		tag.String(),
		expansion.ExpanderRight,
	)

	for id := range seq {
		if err := tags.Add(id); err != nil {
			panic(err)
		}
	}

	clone := CloneTagSet(tags)
	tags.Reset()

	for tag := range WithRemovedCommonPrefixes(clone).All() {
		tags.Add(tag)
	}
}

func RemovePrefixes(haystack TagSetMutable, needle TagStruct) {
	for tag := range haystack.All() {
		if !strings.HasPrefix(tag.String(), needle.String()) {
			continue
		}

		quiter_set.Del(haystack, tag)
	}
}
