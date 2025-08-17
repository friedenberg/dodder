package object_metadata

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/expansion"
	"code.linenisgreat.com/dodder/go/src/charlie/repo_signing"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type Field = string_format_writer.Field

type Metadata struct {
	// Domain
	RepoPubkey repo_signing.PublicKey
	RepoSig    repo_signing.Data
	// InventoryListTai

	Description descriptions.Description
	// TODO refactor this to be an efficient structure backed by a slice
	Tags ids.TagMutableSet // public for gob, but should be private
	Type ids.Type

	Digests
	Tai ids.Tai

	Comments []string
	Cache    Cache
	Fields   []Field
}

func (metadata *Metadata) GetMetadata() *Metadata {
	return metadata
}

func (metadata *Metadata) GetDigest() *sha.Sha {
	return &metadata.SelfMetadataObjectIdParent
}

func (metadata *Metadata) GetMotherDigest() *sha.Sha {
	return &metadata.ParentMetadataObjectIdParent
}

func (metadata *Metadata) UserInputIsEmpty() bool {
	if !metadata.Description.IsEmpty() {
		return false
	}

	if metadata.Tags != nil && metadata.Tags.Len() > 0 {
		return false
	}

	if !ids.IsEmpty(metadata.Type) {
		return false
	}

	return true
}

func (metadata *Metadata) IsEmpty() bool {
	if !metadata.Blob.IsNull() {
		return false
	}

	if !metadata.UserInputIsEmpty() {
		return false
	}

	if !metadata.Tai.IsZero() {
		return false
	}

	return true
}

// TODO fix issue with GetTags being nil sometimes
func (metadata *Metadata) GetTags() ids.TagSet {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	return metadata.Tags
}

func (metadata *Metadata) ResetTags() {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	metadata.Tags.Reset()
	metadata.Cache.TagPaths.Reset()
}

func (metadata *Metadata) AddTagString(tagString string) (err error) {
	if tagString == "" {
		return
	}

	var tag ids.Tag

	if err = tag.Set(tagString); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = metadata.AddTagPtr(&tag); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (metadata *Metadata) AddTagPtr(e *ids.Tag) (err error) {
	if e == nil || e.String() == "" {
		return
	}

	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	ids.AddNormalizedTag(metadata.Tags, e)
	cs := catgut.MakeFromString(e.String())
	metadata.Cache.TagPaths.AddTag(cs)

	return
}

func (metadata *Metadata) AddTagPtrFast(e *ids.Tag) (err error) {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	if err = metadata.Tags.Add(*e); err != nil {
		err = errors.Wrap(err)
		return
	}

	cs := catgut.MakeFromString(e.String())

	if err = metadata.Cache.TagPaths.AddTag(cs); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (metadata *Metadata) SetTags(tags ids.TagSet) {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	metadata.Tags.Reset()

	if tags == nil {
		return
	}

	if tags.Len() == 1 && tags.Any().String() == "" {
		panic("empty tag set")
	}

	for tag := range tags.AllPtr() {
		errors.PanicIfError(metadata.AddTagPtr(tag))
	}
}

func (metadata *Metadata) GetType() ids.Type {
	return metadata.Type
}

func (metadata *Metadata) GetTypePtr() *ids.Type {
	return &metadata.Type
}

func (metadata *Metadata) GetTai() ids.Tai {
	return metadata.Tai
}

// TODO-P2 remove
func (metadata *Metadata) EqualsSansTai(a *Metadata) bool {
	return EqualerSansTai.Equals(a, metadata)
}

// TODO-P2 remove
func (metadata *Metadata) Equals(z1 *Metadata) bool {
	return Equaler.Equals(metadata, z1)
}

func (metadata *Metadata) Subtract(
	b *Metadata,
) {
	if metadata.Type.String() == b.Type.String() {
		metadata.Type = ids.Type{}
	}

	if metadata.Tags == nil {
		return
	}

	// ui.Debug().Print("before", b.Tags, a.Tags)

	for e := range b.Tags.AllPtr() {
		// ui.Debug().Print(e)
		metadata.Tags.DelPtr(e)
	}

	// ui.Debug().Print("after", b.Tags, a.Tags)
}

func (metadata *Metadata) AddComment(f string, vals ...any) {
	metadata.Comments = append(metadata.Comments, fmt.Sprintf(f, vals...))
}

func (metadata *Metadata) SetMutter(mg Getter) (err error) {
	mutter := mg.GetMetadata()

	if err = metadata.GetMotherDigest().SetDigest(
		mutter.GetDigest(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = metadata.ParentMetadataObjectIdParent.SetDigest(
		&mutter.SelfMetadataObjectIdParent,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (metadata *Metadata) GenerateExpandedTags() {
	metadata.Cache.SetExpandedTags(ids.ExpandMany(
		metadata.GetTags(),
		expansion.ExpanderRight,
	))
}

func (metadata *Metadata) GetRepoPubkeyValue() blech32.Value {
	return blech32.Value{
		HRP:  repo_signing.HRPRepoPubKeyV1,
		Data: metadata.RepoPubkey,
	}
}

func (metadata *Metadata) GetRepoSigValue() blech32.Value {
	return blech32.Value{
		HRP:  repo_signing.HRPRepoSigV1,
		Data: metadata.RepoSig,
	}
}
