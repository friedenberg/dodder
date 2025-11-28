package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type tagSetView metadata

// var _ ids.TagSet = (tagSetView)(metadata{})

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

func (tagSet tagSetView) Any() ids.Tag {
	return tagSet.Tags.Any()
}
