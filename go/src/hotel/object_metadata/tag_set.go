package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type (
	tagSetView        metadata
	tagSetViewMutable struct {
		tagSetView
	}
)

var (
	_ ids.TagSet        = (tagSetView)(metadata{})
	_ ids.TagSetMutable = tagSetViewMutable{}
)

func (tagSet tagSetView) Len() int {
	return tagSet.Tags.Len()
}

func (tagSet tagSetView) All() interfaces.Seq[ids.Tag] {
	return func(yield func(ids.Tag) bool) {
		for tag := range tagSet.Tags.All() {
			if !yield(tag) {
				return
			}
		}
	}
}

func (tagSet tagSetView) ContainsKey(key string) bool {
	return tagSet.Tags.ContainsKey(key)
}

func (tagSet tagSetView) Get(key string) (ids.Tag, bool) {
	return tagSet.Tags.Get(key)
}

func (tagSet tagSetView) Key(tag ids.Tag) string {
	return tagSet.Tags.Key(tag)
}

func (tagSet tagSetViewMutable) Add(tag ids.Tag) error {
	return tagSet.Tags.Add(tag)
}

func (tagSet tagSetViewMutable) DelKey(key string) error {
	return tagSet.Tags.DelKey(key)
}

func (tagSet tagSetViewMutable) Reset() {
	tagSet.Tags.Reset()
}
