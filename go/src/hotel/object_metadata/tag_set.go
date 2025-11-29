package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type tagSet struct {
	tags collections_slice.Slice[markl.Lock[ids.Tag, *ids.Tag]]
}

func makeTagSetMutable() ids.TagSetMutable {
	return ids.MakeTagMutableSet()
}

var (
	_ ids.TagSet        = &tagSet{}
	_ ids.TagSetMutable = &tagSet{}
)

func (tagSet tagSet) Len() int {
	return tagSet.tags.Len()
}

func (tagSet tagSet) All() interfaces.Seq[ids.Tag] {
	return func(yield func(ids.Tag) bool) {
		for tag := range tagSet.tags.All() {
			if !yield(tag.GetKey()) {
				return
			}
		}
	}
}

// TODO switch to binary search
func (tagSet tagSet) ContainsKey(key string) bool {
	for tag := range tagSet.tags.All() {
		if tag.GetKey().String() == key {
			return true
		}
	}

	return false
}

// TODO switch to binary search
func (tagSet tagSet) Get(key string) (ids.Tag, bool) {
	for tag := range tagSet.tags.All() {
		if tag.GetKey().String() == key {
			return tag.GetKey(), true
		}
	}

	return ids.Tag{}, false
}

func (tagSet tagSet) Key(tag ids.Tag) string {
	return tag.String()
}

// TODO sort
func (tagSet *tagSet) Add(tag ids.Tag) error {
	tagSet.tags.Append(markl.MakeLockWith(tag, nil))
	return nil
}

func (tagSet tagSet) DelKey(key string) error {
	var found bool
	var index int
	var tagLock TagLock

	for index, tagLock = range tagSet.tags {
		if tagLock.GetKey().String() == key {
			found = true
			break
		}
	}

	if found {
		tagSet.tags.Delete(index, index+1)
	}

	return nil
}

func (tagSet tagSet) Reset() {
	tagSet.tags.Reset()
}
