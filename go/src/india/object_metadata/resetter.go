package object_metadata

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

var Resetter resetter

type resetter struct{}

func (resetter) Reset(metadatuh IMetadataMutable) {
	{
		metadata := metadatuh.(*metadata)
		metadata.Description.Reset()
		metadata.Comments = metadata.Comments[:0]
		metadata.sigRepo.Reset()
		metadata.pubRepo.Reset()
		metadata.ResetTags()
		ResetterIndex.Reset(&metadata.Index)
		metadata.Type = ids.Type{}
		metadata.Tai.Reset()
		metadata.DigBlob.Reset()
		metadata.SelfWithoutTai.Reset()
		metadata.digSelf.Reset()
		metadata.sigMother.Reset()
		metadata.Fields = metadata.Fields[:0]
		metadata.lockfile.tipe.Key = ""
		metadata.lockfile.tipe.Id.Reset()
	}
}

func (resetter) ResetWithExceptFields(dst *metadata, src *metadata) {
	dst.Description = src.Description
	dst.Comments = dst.Comments[:0]
	dst.Comments = append(dst.Comments, src.Comments...)

	dst.SetTagsFast(src.Tags)

	ResetterIndex.ResetWith(&dst.Index, &src.Index)

	dst.sigRepo.ResetWith(src.sigRepo)
	dst.pubRepo.ResetWith(src.pubRepo)

	dst.Type = src.Type
	dst.Tai = src.Tai

	dst.DigBlob.ResetWith(src.DigBlob)
	dst.digSelf.ResetWith(src.digSelf)
	dst.sigMother.ResetWith(src.sigMother)

	dst.lockfile.tipe.Key = src.lockfile.tipe.Key
	dst.lockfile.tipe.Id.ResetWithMarklId(src.lockfile.tipe.Id)
}

func (resetter resetter) ResetWith(dst IMetadataMutable, src IMetadata) {
	{
		dst := dst.(*metadata)
		src := src.(*metadata)
		resetter.ResetWithExceptFields(dst, src)
		dst.Fields = dst.Fields[:0]
		dst.Fields = append(dst.Fields, src.Fields...)
	}
}

var ResetterIndex resetterIndex

type resetterIndex struct{}

func (resetterIndex) Reset(a *Index) {
	a.ParentTai.Reset()
	a.TagPaths.Reset()
	a.Dormant.Reset()
	a.SetExpandedTags(nil)
	a.SetImplicitTags(nil)
	a.QueryPath.Reset()
}

func (resetterIndex) ResetWith(a, b *Index) {
	a.ParentTai.ResetWith(b.ParentTai)
	a.TagPaths.ResetWith(&b.TagPaths)
	a.Dormant.ResetWith(b.Dormant)
	a.SetExpandedTags(b.GetExpandedTags())
	a.SetImplicitTags(b.GetImplicitTags())
	a.QueryPath.Reset()
	a.QueryPath = slices.Grow(a.QueryPath, b.QueryPath.Len())
	copy(a.QueryPath, b.QueryPath)
}
