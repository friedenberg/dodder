package objects

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/tag_paths"
)

type (
	Index interface {
		GetFields() interfaces.Seq[Field]
		GetTagPaths() *tag_paths.Tags // TODO make immutable view
		GetDormant() values.Bool
		GetImplicitTags() TagSet
		GetComments() interfaces.Seq[string]
		GetSelfWithoutTai() interfaces.MarklId
	}

	IndexMutable interface {
		Index

		AddTagsImplicitPtr(tag *Tag) (err error)
		GetDormantMutable() *values.Bool
		GetFieldsMutable() *collections_slice.Slice[Field]
		GetTagPathsMutable() *tag_paths.Tags
		SetImplicitTags(e TagSet)
		GetCommentsMutable() *collections_slice.Slice[string]
		GetSelfWithoutTaiMutable() interfaces.MarklIdMutable
	}
)

type index struct {
	Dormant      values.Bool
	ImplicitTags TagSetMutable // public for gob, but should be private
	TagPaths     tag_paths.Tags
	Comments     collections_slice.Slice[string]
	Fields       collections_slice.Slice[Field]

	keyValues
}

var (
	_ Index        = &index{}
	_ IndexMutable = &index{}
)

func (index *index) GetTagPaths() *tag_paths.Tags {
	return &index.TagPaths
}

func (index *index) GetTagPathsMutable() *tag_paths.Tags {
	return &index.TagPaths
}

func (index *index) GetDormant() values.Bool {
	return index.Dormant
}

func (index *index) GetDormantMutable() *values.Bool {
	return &index.Dormant
}

func (index *index) GetImplicitTags() TagSet {
	return index.GetImplicitTagsMutable()
}

func (index *index) AddTagsImplicitPtr(tag *Tag) (err error) {
	return index.GetImplicitTagsMutable().Add(*tag)
}

func (index *index) GetImplicitTagsMutable() TagSetMutable {
	if index.ImplicitTags == nil {
		index.ImplicitTags = ids.MakeTagSetMutable()
	}

	return index.ImplicitTags
}

func (index *index) SetImplicitTags(tags TagSet) {
	tagsMutable := index.GetImplicitTagsMutable()
	tagsMutable.Reset()

	if tags == nil {
		return
	}

	for tag := range tags.All() {
		errors.PanicIfError(tagsMutable.Add(tag))
	}
}

func (index *index) GetComments() interfaces.Seq[string] {
	return index.Comments.All()
}

func (index *index) GetCommentsMutable() *collections_slice.Slice[string] {
	return &index.Comments
}

func (metadata *metadata) AddComment(f string, vals ...any) {
	metadata.Index.Comments = append(metadata.Index.Comments, fmt.Sprintf(f, vals...))
}

func (index *index) GetFields() interfaces.Seq[Field] {
	return index.Fields.All()
}

func (index *index) GetFieldsMutable() *collections_slice.Slice[Field] {
	return &index.Fields
}
