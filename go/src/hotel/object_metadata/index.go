package object_metadata

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/golf/tag_paths"
)

type (
	IIndex interface {
		GetFields() interfaces.Seq[Field]
		GetTagPaths() *tag_paths.Tags // TODO make immutable view
		GetExpandedTags() ids.TagSet
		GetDormant() values.Bool
		GetImplicitTags() ids.TagSet
		GetParentTai() ids.Tai
		GetComments() interfaces.Seq[string] // TODO move to IIndex
	}

	IIndexMutable interface {
		IIndex

		AddTagExpandedPtr(e *ids.Tag) (err error)
		AddTagsImplicitPtr(tag *ids.Tag) (err error)
		GetDormantMutable() *values.Bool
		GetExpandedTagsMutable() ids.TagMutableSet
		GetFieldsMutable() *collections_slice.Slice[Field]
		GetParentTaiMutable() *ids.Tai
		GetTagPathsMutable() *tag_paths.Tags
		SetExpandedTags(tags ids.TagSet)
		SetImplicitTags(e ids.TagSet)
		GetCommentsMutable() *collections_slice.Slice[string] // TODO move to IIndexMutable
	}
)

type Index struct {
	ParentTai    ids.Tai // TODO remove in favor of MotherSig
	Dormant      values.Bool
	ExpandedTags ids.TagMutableSet // public for gob, but should be private
	ImplicitTags ids.TagMutableSet // public for gob, but should be private
	TagPaths     tag_paths.Tags
	Comments     collections_slice.Slice[string]
	Fields       collections_slice.Slice[Field]

	QueryPath

	keyValues
}

var (
	_ IIndex        = &Index{}
	_ IIndexMutable = &Index{}
)

func (index *Index) GetTagPaths() *tag_paths.Tags {
	return &index.TagPaths
}

func (index *Index) GetTagPathsMutable() *tag_paths.Tags {
	return &index.TagPaths
}

func (index *Index) GetDormant() values.Bool {
	return index.Dormant
}

func (index *Index) GetDormantMutable() *values.Bool {
	return &index.Dormant
}

func (index *Index) GetExpandedTags() ids.TagSet {
	return index.GetExpandedTagsMutable()
}

func (index *Index) AddTagExpandedPtr(e *ids.Tag) (err error) {
	return quiter.AddClonePool(
		index.GetExpandedTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		e,
	)
}

func (index *Index) GetExpandedTagsMutable() ids.TagMutableSet {
	if index.ExpandedTags == nil {
		index.ExpandedTags = ids.MakeTagMutableSet()
	}

	return index.ExpandedTags
}

func (index *Index) SetExpandedTags(tags ids.TagSet) {
	tagsExpanded := index.GetExpandedTagsMutable()
	quiter.ResetMutableSetWithPool(tagsExpanded, ids.GetTagPool())

	if tags == nil {
		return
	}

	for tag := range tags.All() {
		errors.PanicIfError(tagsExpanded.Add(tag))
	}
}

func (index *Index) GetImplicitTags() ids.TagSet {
	return index.GetImplicitTagsMutable()
}

func (index *Index) AddTagsImplicitPtr(tag *ids.Tag) (err error) {
	return quiter.AddClonePool(
		index.GetImplicitTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		tag,
	)
}

func (index *Index) GetImplicitTagsMutable() ids.TagMutableSet {
	if index.ImplicitTags == nil {
		index.ImplicitTags = ids.MakeTagMutableSet()
	}

	return index.ImplicitTags
}

func (index *Index) SetImplicitTags(tags ids.TagSet) {
	tagsMutable := index.GetImplicitTagsMutable()
	quiter.ResetMutableSetWithPool(tagsMutable, ids.GetTagPool())

	if tags == nil {
		return
	}

	for tag := range tags.All() {
		errors.PanicIfError(tagsMutable.Add(tag))
	}
}

func (index *Index) GetParentTai() ids.Tai {
	return index.ParentTai
}

func (index *Index) GetParentTaiMutable() *ids.Tai {
	return &index.ParentTai
}

func (index *Index) GetComments() interfaces.Seq[string] {
	return index.Comments.All()
}

func (index *Index) GetCommentsMutable() *collections_slice.Slice[string] {
	return &index.Comments
}

func (metadata *metadata) AddComment(f string, vals ...any) {
	metadata.Index.Comments = append(metadata.Index.Comments, fmt.Sprintf(f, vals...))
}

func (index *Index) GetFields() interfaces.Seq[Field] {
	return index.Fields.All()
}

func (index *Index) GetFieldsMutable() *collections_slice.Slice[Field] {
	return &index.Fields
}
