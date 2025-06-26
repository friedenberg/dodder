package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/tag_paths"
)

type Cache struct {
	ParentTai    ids.Tai
	Dormant      values.Bool
	ExpandedTags ids.TagMutableSet // public for gob, but should be private
	ImplicitTags ids.TagMutableSet // public for gob, but should be private
	TagPaths     tag_paths.Tags
	QueryPath
}

func (v *Cache) GetExpandedTags() ids.TagSet {
	return v.GetExpandedTagsMutable()
}

func (v *Cache) AddTagExpandedPtr(e *ids.Tag) (err error) {
	return quiter.AddClonePool(
		v.GetExpandedTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		e,
	)
}

func (v *Cache) GetExpandedTagsMutable() ids.TagMutableSet {
	if v.ExpandedTags == nil {
		v.ExpandedTags = ids.MakeTagMutableSet()
	}

	return v.ExpandedTags
}

func (v *Cache) SetExpandedTags(e ids.TagSet) {
	es := v.GetExpandedTagsMutable()
	quiter.ResetMutableSetWithPool(es, ids.GetTagPool())

	if e == nil {
		return
	}

	errors.PanicIfError(e.Each(es.Add))
}

func (v *Cache) GetImplicitTags() ids.TagSet {
	return v.GetImplicitTagsMutable()
}

func (v *Cache) AddTagsImplicitPtr(e *ids.Tag) (err error) {
	return quiter.AddClonePool(
		v.GetImplicitTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		e,
	)
}

func (v *Cache) GetImplicitTagsMutable() ids.TagMutableSet {
	if v.ImplicitTags == nil {
		v.ImplicitTags = ids.MakeTagMutableSet()
	}

	return v.ImplicitTags
}

func (v *Cache) SetImplicitTags(e ids.TagSet) {
	es := v.GetImplicitTagsMutable()
	quiter.ResetMutableSetWithPool(es, ids.GetTagPool())

	if e == nil {
		return
	}

	errors.PanicIfError(e.Each(es.Add))
}
