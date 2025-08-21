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

func (cache *Cache) GetExpandedTags() ids.TagSet {
	return cache.GetExpandedTagsMutable()
}

func (cache *Cache) AddTagExpandedPtr(e *ids.Tag) (err error) {
	return quiter.AddClonePool(
		cache.GetExpandedTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		e,
	)
}

func (cache *Cache) GetExpandedTagsMutable() ids.TagMutableSet {
	if cache.ExpandedTags == nil {
		cache.ExpandedTags = ids.MakeTagMutableSet()
	}

	return cache.ExpandedTags
}

func (cache *Cache) SetExpandedTags(tags ids.TagSet) {
	tagsExpanded := cache.GetExpandedTagsMutable()
	quiter.ResetMutableSetWithPool(tagsExpanded, ids.GetTagPool())

	if tags == nil {
		return
	}

	for tag := range tags.All() {
		errors.PanicIfError(tagsExpanded.Add(tag))
	}
}

func (cache *Cache) GetImplicitTags() ids.TagSet {
	return cache.GetImplicitTagsMutable()
}

func (cache *Cache) AddTagsImplicitPtr(tag *ids.Tag) (err error) {
	return quiter.AddClonePool(
		cache.GetImplicitTagsMutable(),
		ids.GetTagPool(),
		ids.TagResetter,
		tag,
	)
}

func (cache *Cache) GetImplicitTagsMutable() ids.TagMutableSet {
	if cache.ImplicitTags == nil {
		cache.ImplicitTags = ids.MakeTagMutableSet()
	}

	return cache.ImplicitTags
}

func (cache *Cache) SetImplicitTags(e ids.TagSet) {
	es := cache.GetImplicitTagsMutable()
	quiter.ResetMutableSetWithPool(es, ids.GetTagPool())

	if e == nil {
		return
	}

	for tag := range e.All() {
		errors.PanicIfError(es.Add(tag))
	}
}
