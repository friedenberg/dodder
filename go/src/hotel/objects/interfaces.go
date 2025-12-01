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

	TagLock        = interfaces.Lock[ids.Tag, *ids.Tag]
	TagLockMutable = interfaces.LockMutable[ids.Tag, *ids.Tag]

	TagSetMutable interface {
		ids.TagSet
	}

	Metadata interface {
		Getter

		IsEmpty() bool

		GetDescription() descriptions.Description
		GetIndex() Index
		GetTags() ids.TagSet
		AllTags() interfaces.Seq[ids.Tag]
		GetTai() ids.Tai
		GetType() ids.Type
		GetTypeLock() TypeLock

		GetTagLock(ids.Tag) TagLock

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

		// TODO rewrite
		AddTagPtr(e *ids.Tag) (err error)
		AddTag(ids.Tag) (err error)
		ResetTags()
		SetTags(ids.TagSet)
		AddTagString(tagString string) (err error)
		AddTagPtrFast(tag *ids.Tag) (err error)
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
		GetTagLockMutable(ids.Tag) TagLockMutable
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
