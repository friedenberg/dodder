package ids

import (
	_ "encoding/gob"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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

func TagStructSeqToITag(tags interfaces.Seq[TagStruct]) interfaces.Seq[ITag] {
	return func(yield func(ITag) bool) {
		for tag := range tags {
			if !yield(tag) {
				return
			}
		}
	}
}

func ITagSeqToTagStructSeq(itags interfaces.Seq[ITag]) interfaces.Seq[TagStruct] {
	return func(yield func(Tag) bool) {
		for itag := range itags {
			var tag TagStruct

			errors.PanicIfError(tag.Set(itag.String()))

			if !yield(tag) {
				return
			}
		}
	}
}

func ITagSeqToTagStructPtrSeq(itags interfaces.Seq[ITag]) interfaces.Seq[*TagStruct] {
	return func(yield func(*TagStruct) bool) {
		for itag := range itags {
			var tag TagStruct

			errors.PanicIfError(tag.Set(itag.String()))

			if !yield(&tag) {
				return
			}
		}
	}
}

func MakeTagSetFromISeq(itags interfaces.Seq[ITag]) (s TagSet) {
	tags := ITagSeqToTagStructPtrSeq(itags)
	return collections_ptr.MakeValueSetSeq(nil, tags, 0)
}

func MakeTagSetFromISlice(itags ...ITag) (s TagSet) {
	if len(itags) == 0 {
		return TagSetEmpty
	}

	tags := func(yield func(*Tag) bool) {
		for _, itag := range itags {
			var tag TagStruct

			errors.PanicIfError(tag.Set(itag.String()))

			if !yield(&tag) {
				return
			}
		}
	}

	return collections_ptr.MakeValueSetSeq(nil, tags, len(itags))
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

func SubtractPrefix(input TagSet, tag TagStruct) (output TagSet) {
	s3 := MakeTagSetMutable()

	for _, e1 := range quiter.CollectSlice(input) {
		e2, _ := LeftSubtract(e1, tag)

		if e2.String() == "" {
			continue
		}

		s3.Add(e2)
	}

	output = CloneTagSet(s3)

	return output
}

func TagSetMutableAdd(set TagSetMutable, itag ITag) {
	var tag TagStruct
	errors.PanicIfError(tag.Set(itag.String()))
	errors.PanicIfError(set.Add(tag))
}

func TagEquals(left, right ITag) bool {
	return left.String() == right.String()
}
