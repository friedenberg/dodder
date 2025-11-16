package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
)

type (
	IIndex interface {
		GetTagPaths() *tag_paths.Tags // TODO make immutable view
		GetExpandedTags() ids.TagSet
		GetDormant() values.Bool
		GetImplicitTags() ids.TagSet
	}

	IIndexMutable interface {
		IIndex

		GetDormantMutable() *values.Bool
		AddTagExpandedPtr(e *ids.Tag) (err error)
		GetExpandedTagsMutable() ids.TagMutableSet
		SetExpandedTags(tags ids.TagSet)
		SetImplicitTags(e ids.TagSet)
	}
)

type Index struct {
	ParentTai    ids.Tai // TODO remove in favor of MotherSig
	Dormant      values.Bool
	ExpandedTags ids.TagMutableSet // public for gob, but should be private
	ImplicitTags ids.TagMutableSet // public for gob, but should be private
	TagPaths     tag_paths.Tags
	QueryPath
}

var (
	_ IIndex        = &Index{}
	_ IIndexMutable = &Index{}
)

func (index *Index) GetTagPaths() *tag_paths.Tags {
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

func (index *Index) SetImplicitTags(e ids.TagSet) {
	es := index.GetImplicitTagsMutable()
	quiter.ResetMutableSetWithPool(es, ids.GetTagPool())

	if e == nil {
		return
	}

	for tag := range e.All() {
		errors.PanicIfError(es.Add(tag))
	}
}
