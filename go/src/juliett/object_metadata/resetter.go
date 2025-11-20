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
		metadata.sigRepo.Reset()
		metadata.pubRepo.Reset()
		metadata.ResetTags()
		resetIndex(&metadata.Index)
		metadata.Type = ids.Type{}
		metadata.Tai.Reset()
		metadata.DigBlob.Reset()
		metadata.digSelf.Reset()
		metadata.sigMother.Reset()
		metadata.lockfile.tipe.Key = ""
		metadata.lockfile.tipe.Id.Reset()
	}
}

func (resetter) ResetWithExceptFields(dst *metadata, src *metadata) {
	dst.Description = src.Description

	dst.SetTagsFast(src.Tags)

	resetIndexWith(&dst.Index, &src.Index)

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
		dst.Index.Fields.ResetWith(src.Index.Fields)
	}
}

func resetIndex(a *Index) {
	a.ParentTai.Reset()
	a.TagPaths.Reset()
	a.Dormant.Reset()
	a.SetExpandedTags(nil)
	a.SetImplicitTags(nil)
	a.QueryPath.Reset()
	a.Comments = a.Comments[:0]
	a.SelfWithoutTai.Reset()
}

func resetIndexWith(dst, src *Index) {
	dst.ParentTai.ResetWith(src.ParentTai)
	dst.TagPaths.ResetWith(&src.TagPaths)
	dst.Dormant.ResetWith(src.Dormant)
	dst.SetExpandedTags(src.GetExpandedTags())
	dst.SetImplicitTags(src.GetImplicitTags())
	dst.QueryPath.Reset()
	dst.QueryPath = slices.Grow(dst.QueryPath, src.QueryPath.Len())
	copy(dst.QueryPath, src.QueryPath)

	dst.Comments = dst.Comments[:0]
	dst.Comments = append(dst.Comments, src.Comments...)
}
