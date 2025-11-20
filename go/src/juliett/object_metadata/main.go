package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/catgut"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type Field = string_format_writer.Field

// TODO transform into a view interface that can be backed by various
// representations
type metadata struct {
	Description descriptions.Description
	// TODO refactor this to be an efficient structure backed by a slice
	Tags ids.TagMutableSet // public for gob, but should be private
	Type ids.Type

	DigBlob   markl.Id
	digSelf   markl.Id
	pubRepo   markl.Id
	sigMother markl.Id
	sigRepo   markl.Id

	Tai ids.Tai

	Index Index

	lockfile
}

var (
	_ IMetadata        = &metadata{}
	_ IMetadataMutable = &metadata{}
	_ Getter           = &metadata{}
	_ GetterMutable    = &metadata{}
)

func (metadata *metadata) GetMetadata() IMetadata {
	return metadata
}

func (metadata *metadata) GetMetadataMutable() IMetadataMutable {
	return metadata
}

func (metadata *metadata) GetIndex() IIndex {
	return &metadata.Index
}

func (metadata *metadata) GetIndexMutable() IIndexMutable {
	return &metadata.Index
}

func (metadata *metadata) GetDescription() descriptions.Description {
	return metadata.Description
}

func (metadata *metadata) GetDescriptionMutable() *descriptions.Description {
	return &metadata.Description
}

func (metadata *metadata) GetTai() ids.Tai {
	return metadata.Tai
}

func (metadata *metadata) GetTaiMutable() *ids.Tai {
	return &metadata.Tai
}

func (metadata *metadata) GetLockfile() Lockfile {
	return metadata.lockfile
}

func (metadata *metadata) GetLockfileMutable() LockfileMutable {
	return &metadata.lockfile
}

func (metadata *metadata) UserInputIsEmpty() bool {
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

func (metadata *metadata) IsEmpty() bool {
	if !metadata.DigBlob.IsNull() {
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
func (metadata *metadata) GetTags() ids.TagSet {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	return metadata.Tags
}

func (metadata *metadata) ResetTags() {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	metadata.Tags.Reset()
	metadata.Index.TagPaths.Reset()
}

func (metadata *metadata) AddTagString(tagString string) (err error) {
	if tagString == "" {
		return err
	}

	var tag ids.Tag

	if err = tag.Set(tagString); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = metadata.AddTagPtr(&tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (metadata *metadata) AddTagPtr(e *ids.Tag) (err error) {
	if e == nil || e.String() == "" {
		return err
	}

	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	ids.AddNormalizedTag(metadata.Tags, e)
	cs := catgut.MakeFromString(e.String())
	metadata.Index.TagPaths.AddTag(cs)

	return err
}

func (metadata *metadata) AddTagPtrFast(tag *ids.Tag) (err error) {
	if metadata.Tags == nil {
		metadata.Tags = ids.MakeTagMutableSet()
	}

	if err = metadata.Tags.Add(*tag); err != nil {
		err = errors.Wrap(err)
		return err
	}

	tagBytestring := catgut.MakeFromString(tag.String())

	if err = metadata.Index.TagPaths.AddTag(tagBytestring); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (metadata *metadata) SetTags(tags ids.TagSet) {
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

func (metadata *metadata) SetTagsFast(tags ids.TagSet) {
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
		errors.PanicIfError(metadata.AddTagPtrFast(tag))
	}
}

func (metadata *metadata) GetType() ids.Type {
	return metadata.Type
}

func (metadata *metadata) GetTypeMutable() *ids.Type {
	return &metadata.Type
}

func (metadata *metadata) Subtract(
	b *metadata,
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

func (metadata *metadata) GenerateExpandedTags() {
	metadata.Index.SetExpandedTags(ids.ExpandMany(
		metadata.GetTags(),
		expansion.ExpanderRight,
	))
}
