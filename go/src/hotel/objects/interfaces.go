package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type (
	MetadataStruct = metadata

	TagStruct     = ids.TagStruct
	Tag           = ids.Tag
	TagSet        = ids.Set[ids.TagStruct]
	TagSetMutable = ids.SetMutable[ids.TagStruct]

	IdLock        = interfaces.Lock[SeqId, *SeqId]
	IdLockMutable = interfaces.LockMutable[SeqId, *SeqId]

	TypeLock        = interfaces.Lock[Type, TypeMutable]
	TypeLockMutable = interfaces.LockMutable[Type, TypeMutable]
	TagLock         = IdLock
	TagLockMutable  = IdLockMutable

	Type        = ids.SeqId
	TypeMutable = *ids.SeqId

	Metadata interface {
		Getter

		IsEmpty() bool

		GetDescription() descriptions.Description
		GetIndex() Index
		GetTags() TagSet
		AllTags() interfaces.Seq[Tag]
		GetTai() ids.Tai
		GetType() Type
		GetTypeLock() TypeLock

		GetTagLock(Tag) TagLock

		GetBlobDigest() interfaces.MarklId
		GetObjectDigest() interfaces.MarklId
		GetMotherObjectSig() interfaces.MarklId
		GetRepoPubKey() interfaces.MarklId
		GetObjectSig() interfaces.MarklId
	}

	MetadataMutable interface {
		interfaces.CommandComponentWriter
		Metadata
		GetterMutable

		Subtract(Metadata)

		AddTagPtr(Tag) (err error)
		ResetTags()
		AddTagString(tagString string) (err error)
		AddTagPtrFast(tag Tag) (err error)
		GenerateExpandedTags()

		GetIndexMutable() IndexMutable

		GetBlobDigestMutable() interfaces.MarklIdMutable
		GetDescriptionMutable() *descriptions.Description
		GetMotherObjectSigMutable() interfaces.MarklIdMutable
		GetObjectDigestMutable() interfaces.MarklIdMutable
		GetObjectSigMutable() interfaces.MarklIdMutable
		GetRepoPubKeyMutable() interfaces.MarklIdMutable
		GetTaiMutable() *ids.Tai
		GetTypeMutable() TypeMutable
		GetTypeLockMutable() TypeLockMutable
		GetTagLockMutable(Tag) TagLockMutable
	}

	Getter interface {
		GetMetadata() Metadata
	}

	GetterMutable interface {
		GetMetadataMutable() MetadataMutable
	}

	PersistentFormatterContext interface {
		Getter
		// TODO determine if this is necessary
		GetterMutable
		// GetObjectId() ids.Id
	}

	PersistentParserContext interface {
		GetterMutable
		// GetObjectIdMutable() ids.IdMutable
	}
)
