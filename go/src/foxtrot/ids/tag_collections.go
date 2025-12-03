package ids

import (
	_ "encoding/gob"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
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

func IntersectPrefixes(haystack TagSet, needle TagStruct) (s3 TagSet) {
	s4 := MakeTagSetMutable()

	for _, e := range quiter.CollectSlice(haystack) {
		if strings.HasPrefix(e.String(), needle.String()) {
			s4.Add(e)
		}
	}

	s3 = CloneTagSet(s4)

	return s3
}

func SubtractPrefix(s1 TagSet, e TagStruct) (s2 TagSet) {
	s3 := MakeTagSetMutable()

	for _, e1 := range quiter.CollectSlice(s1) {
		e2, _ := LeftSubtract(e1, e)

		if e2.String() == "" {
			continue
		}

		s3.Add(e2)
	}

	s2 = CloneTagSet(s3)

	return s2
}
