package object_metadata

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
		metadata.Type.Reset()
		metadata.Tai.Reset()
		metadata.DigBlob.Reset()
		metadata.digSelf.Reset()
		metadata.sigMother.Reset()
	}
}

func (resetter) ResetWithExceptFields(dst IMetadataMutable, src IMetadata) {
	{
		dst := dst.(*metadata)
		src := src.(*metadata)

		dst.Description = src.Description

		dst.SetTagsFast(src.Tags)

		resetIndexWith(&dst.Index, &src.Index)

		dst.sigRepo.ResetWith(src.sigRepo)
		dst.pubRepo.ResetWith(src.pubRepo)

		dst.Type.ResetWith(src.Type)
		dst.Tai = src.Tai

		dst.DigBlob.ResetWith(src.DigBlob)
		dst.digSelf.ResetWith(src.digSelf)
		dst.sigMother.ResetWith(src.sigMother)
	}
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
	a.SetImplicitTags(nil)
	a.Comments.Reset()
	a.SelfWithoutTai.Reset()
}

func resetIndexWith(dst, src *Index) {
	dst.ParentTai.ResetWith(src.ParentTai)
	dst.TagPaths.ResetWith(&src.TagPaths)
	dst.Dormant.ResetWith(src.Dormant)
	dst.SetImplicitTags(src.GetImplicitTags())
	dst.Comments.ResetWith(src.Comments)
}
