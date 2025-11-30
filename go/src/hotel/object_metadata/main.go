package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/expansion"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter_set"
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
	Tags tagSet // public for gob, but should be private
	Type markl.Lock[ids.Type, *ids.Type]

	DigBlob   markl.Id
	digSelf   markl.Id
	pubRepo   markl.Id
	sigMother markl.Id
	sigRepo   markl.Id

	Tai ids.Tai

	Index Index
}

var (
	_ IMetadata        = &metadata{}
	_ IMetadataMutable = &metadata{}
	_ Getter           = &metadata{}
	_ GetterMutable    = &metadata{}
)

func Make() *metadata {
	metadata := &metadata{}
	Resetter.Reset(metadata)
	return metadata
}

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

func (metadata *metadata) UserInputIsEmpty() bool {
	if !metadata.Description.IsEmpty() {
		return false
	}

	if metadata.Tags.Len() > 0 {
		return false
	}

	if !ids.IsEmpty(metadata.GetType()) {
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
	return metadata.Tags
}

func (metadata *metadata) AllTags() interfaces.Seq[ids.Tag] {
	return func(yield func(ids.Tag) bool) {
		for tag := range metadata.Tags.All() {
			if !yield(tag) {
				return
			}
		}
	}
}

func (metadata *metadata) ResetTags() {
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

func (metadata *metadata) AddTag(tag ids.Tag) (err error) {
	return metadata.AddTagPtr(&tag)
}

func (metadata *metadata) AddTagPtr(tag *ids.Tag) (err error) {
	if tag == nil || tag.IsEmpty() {
		return err
	}

	ids.AddNormalizedTag(&metadata.Tags, tag)
	cs := catgut.MakeFromString(tag.String())
	metadata.Index.TagPaths.AddTag(cs)

	return err
}

func (metadata *metadata) AddTagPtrFast(tag *ids.Tag) (err error) {
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
	metadata.Tags.Reset()

	if tags == nil {
		return
	}

	if tags.Len() == 1 && quiter_set.Any(tags).String() == "" {
		panic("empty tag set")
	}

	for tag := range tags.All() {
		errors.PanicIfError(metadata.AddTagPtr(&tag))
	}
}

func (metadata *metadata) SetTagsFast(tags ids.TagSet) {
	metadata.Tags.Reset()

	if tags == nil {
		return
	}

	if tags.Len() == 1 && quiter_set.Any(tags).String() == "" {
		panic("empty tag set")
	}

	for tag := range tags.All() {
		errors.PanicIfError(metadata.AddTagPtrFast(&tag))
	}
}

func (metadata *metadata) GetType() ids.Type {
	return metadata.Type.Key
}

func (metadata *metadata) GetTypeMutable() *ids.Type {
	return &metadata.Type.Key
}

func (metadata *metadata) GetTypeLock() TypeLock {
	return &metadata.Type
}

func (metadata *metadata) GetTypeLockMutable() TypeLockMutable {
	return &metadata.Type
}

func (metadata *metadata) GetTagLock(tag ids.Tag) TagLock {
	lock, _ := metadata.Tags.getLock(tag.String())
	return lock
}

func (metadata *metadata) GetTagLockMutable(tag ids.Tag) TagLockMutable {
	lock, _ := metadata.Tags.getLockMutable(tag.String())
	return lock
}

func (metadata *metadata) Subtract(otherMetadata IMetadata) {
	if metadata.GetType().String() == otherMetadata.GetType().String() {
		metadata.GetTypeMutable().Reset()
	}

	for tag := range otherMetadata.AllTags() {
		quiter_set.Del(&metadata.Tags, tag)
	}
}

func (metadata *metadata) GenerateExpandedTags() {
	metadata.Index.SetExpandedTags(ids.ExpandMany(
		metadata.GetTags().All(),
		expansion.ExpanderRight,
	))
}
