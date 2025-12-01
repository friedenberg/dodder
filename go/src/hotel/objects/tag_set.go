package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	tagLock = markl.Lock[ids.Tag, *ids.Tag]

	tagStruct struct {
		// TODO add path information

		// required to be exported for Gob's stupid illusions
		Lock tagLock
	}

	tagSet struct {
		// required to be exported for Gob's stupid illusions
		// TODO refactor this to use binary searches
		Tags collections_slice.Slice[tagStruct]
	}
)

var (
	_ ids.TagSet        = &tagSet{}
	_ ids.TagSetMutable = &tagSet{}
)

func (tag tagStruct) GetKey() ids.Tag {
	return tag.Lock.GetKey()
}

func (tagSet tagSet) Len() int {
	return tagSet.Tags.Len()
}

func (tagSet tagSet) All() interfaces.Seq[ids.Tag] {
	return func(yield func(ids.Tag) bool) {
		for tag := range tagSet.Tags.All() {
			if !yield(tag.GetKey()) {
				return
			}
		}
	}
}

// TODO switch to binary search
func (tagSet tagSet) ContainsKey(key string) bool {
	for tag := range tagSet.Tags.All() {
		if tag.GetKey().String() == key {
			return true
		}
	}

	return false
}

func (tagSet tagSet) getLock(key string) (TagLock, bool) {
	for tag := range tagSet.Tags.All() {
		if tag.GetKey().String() == key {
			return tag.Lock, true
		}
	}

	return nil, false
}

func (tagSet tagSet) getLockMutable(key string) (TagLockMutable, bool) {
	for index := range tagSet.Tags {
		tag := &tagSet.Tags[index]

		if tag.GetKey().String() == key {
			return &tag.Lock, true
		}
	}

	return nil, false
}

// TODO switch to binary search
func (tagSet tagSet) Get(key string) (ids.Tag, bool) {
	for tag := range tagSet.Tags.All() {
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
	if _, alreadyExists := tagSet.Get(tag.String()); alreadyExists {
		return nil
	}

	tagSet.Tags.Append(tagStruct{
		Lock: markl.MakeLockWith(tag, nil),
	})

	return nil
}

func (tagSet *tagSet) DelKey(key string) error {
	var found bool
	var index int
	var tag tagStruct

	for index, tag = range tagSet.Tags {
		if tag.GetKey().String() == key {
			found = true
			break
		}
	}

	if found {
		tagSet.Tags.Delete(index, index+1)
	}

	return nil
}

func (tagSet *tagSet) Reset() {
	tagSet.Tags.Reset()
}
