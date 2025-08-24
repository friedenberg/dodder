package object_metadata

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

var Resetter resetter

type resetter struct{}

func (resetter) Reset(metadata *Metadata) {
	metadata.Description.Reset()
	metadata.Comments = metadata.Comments[:0]
	metadata.RepoSig = nil
	metadata.RepoPubkey = nil
	metadata.ResetTags()
	ResetterCache.Reset(&metadata.Cache)
	metadata.Type = ids.Type{}
	metadata.Tai.Reset()
	metadata.Blob.Reset()
	metadata.SelfWithoutTai.Reset()
	metadata.self.Reset()
	metadata.mother.Reset()
	metadata.Fields = metadata.Fields[:0]
}

func (resetter) ResetWithExceptFields(dst *Metadata, src *Metadata) {
	dst.Description = src.Description
	dst.Comments = dst.Comments[:0]
	dst.Comments = append(dst.Comments, src.Comments...)

	dst.SetTagsFast(src.Tags)

	ResetterCache.ResetWith(&dst.Cache, &src.Cache)

	dst.RepoSig.ResetWith(src.RepoSig)
	dst.RepoPubkey.ResetWith(src.RepoPubkey)

	dst.Type = src.Type
	dst.Tai = src.Tai

	dst.Blob.ResetWith(&src.Blob)
	dst.self.ResetWith(
		&src.self,
	)
	dst.mother.ResetWith(
		&src.mother,
	)
}

func (r resetter) ResetWith(dst *Metadata, src *Metadata) {
	r.ResetWithExceptFields(dst, src)
	dst.Fields = dst.Fields[:0]
	dst.Fields = append(dst.Fields, src.Fields...)
}

var ResetterCache resetterCache

type resetterCache struct{}

func (resetterCache) Reset(a *Index) {
	a.ParentTai.Reset()
	a.TagPaths.Reset()
	a.Dormant.Reset()
	a.SetExpandedTags(nil)
	a.SetImplicitTags(nil)
	a.QueryPath.Reset()
}

func (resetterCache) ResetWith(a, b *Index) {
	a.ParentTai.ResetWith(b.ParentTai)
	a.TagPaths.ResetWith(&b.TagPaths)
	a.Dormant.ResetWith(b.Dormant)
	a.SetExpandedTags(b.GetExpandedTags())
	a.SetImplicitTags(b.GetImplicitTags())
	a.QueryPath.Reset()
	a.QueryPath = slices.Grow(a.QueryPath, b.QueryPath.Len())
	copy(a.QueryPath, b.QueryPath)
}
