package objects

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
)

type (
	MetadataStruct = metadata

	TypeLock        = interfaces.Lock[ids.Type, *ids.Type]
	TypeLockMutable = interfaces.LockMutable[ids.Type, *ids.Type]

	Tag           = ids.TagStruct
	TagSet        = ids.Set[ids.Tag]
	TagSetMutable = ids.SetMutable[ids.Tag]

	TagLock        = interfaces.Lock[Tag, *Tag]
	TagLockMutable = interfaces.LockMutable[Tag, *Tag]

	IdLock        = interfaces.Lock[SeqId, *SeqId]
	IdLockMutable = interfaces.LockMutable[SeqId, *SeqId]

	Metadata interface {
		Getter

		IsEmpty() bool

		GetDescription() descriptions.Description
		GetIndex() Index
		GetTags() TagSet
		AllTags() interfaces.Seq[Tag]
		GetTai() ids.Tai
		GetType() ids.Type
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
		GetTypeMutable() *ids.Type
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
		GetterMutable
	}

	PersistentParserContext interface {
		GetterMutable
	}
)
