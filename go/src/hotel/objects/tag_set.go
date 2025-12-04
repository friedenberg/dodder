package objects

import (
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	tagLock = markl.Lock[Tag, *Tag]

	tagSet struct {
		// required to be exported for Gob's stupid illusions
		// TODO refactor this to use binary searches
		Tags collections_slice.Slice[tagStruct]
	}
)

var (
	_ TagSet        = &tagSet{}
	_ TagSetMutable = &tagSet{}
)

func (tag tagStruct) GetKey() Tag {
	return tag.Lock.GetKey()
}

func (tagSet tagSet) Len() int {
	return tagSet.Tags.Len()
}

func (tagSet tagSet) All() interfaces.Seq[Tag] {
	return func(yield func(Tag) bool) {
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
func (tagSet tagSet) Get(key string) (Tag, bool) {
	for tag := range tagSet.Tags.All() {
		if tag.GetKey().String() == key {
			return tag.GetKey(), true
		}
	}

	return Tag{}, false
}

func (tagSet tagSet) Key(tag Tag) string {
	return tag.String()
}

// TODO sort
func (tagSet *tagSet) Add(tag Tag) error {
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

func (tagSet *tagSet) addNormalizedTag(tag ITag) {
	seq := expansion.ExpandOneIntoIds[Tag](
		tag.String(),
		expansion.ExpanderRight,
	)

	for id := range seq {
		errors.PanicIfError(tagSet.Add(id))
	}

	sorted := quiter.SortedValuesBy(
		tagSet.Tags,
		tagStructCompareTagKey,
	)

	var lastTag *tagStruct

	for index := range sorted {
		tag := &sorted[index]

		if index == 0 {
			// no need to do anything, this is the first
			lastTag = tag
			continue
		}

		tagString := tag.Lock.GetKey().String()
		lastTagString := lastTag.Lock.GetKey().String()

		switch {
		case strings.HasPrefix(lastTagString, tagString):
			continue

			// replace the shorter value with the longer value that contains the
			// shorter value
		case strings.HasPrefix(tagString, lastTagString):
			if lastTag.Lock.Value.IsEmpty() {
				tagSet.DelKey(lastTagString)
			}

		default:
			// keep the tag
		}

		lastTag = tag
	}
}
