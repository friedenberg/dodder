package ids

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
)

type (
	TagSlice = collections_slice.Slice[Tag]

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
	TagSetEmpty = collections_ptr.MakeValueSetValue[Tag](nil)
}

// TODO move to quiter
func CloneTagSet(tags TagSet) TagSet {
	clone := MakeMutableTagSet()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
}

// TODO move to quiter
func CloneTagSetMutable(tags TagSet) TagSetMutable {
	clone := MakeMutableTagSet()

	for tag := range tags.All() {
		clone.Add(tag)
	}

	return clone
}

// TODO move to quiter
func MakeTagSetFromSlice(tags ...Tag) (s TagSet) {
	if len(tags) == 0 {
		return TagSetEmpty
	}

	return collections_ptr.MakeValueSetValue(nil, tags...)
}

// TODO move to quiter
func MakeTagSetStrings(tagStrings ...string) (s TagSet, err error) {
	return collections_ptr.MakeValueSetString[Tag](nil, tagStrings...)
}

// TODO move to quiter
func MakeMutableTagSet(tags ...Tag) TagSetMutable {
	return MakeTagMutableSet(tags...)
}

// TODO move to quiter
func MakeTagMutableSet(tags ...Tag) TagSetMutable {
	return collections_ptr.MakeMutableValueSetValue(nil, tags...)
}
