package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	Metadata = metadata

	TypeLock = markl.Lock[ids.Type, *ids.Type]

	IMetadata interface {
		Getter

		IsEmpty() bool

		GetDescription() descriptions.Description
		GetIndex() IIndex
		GetTags() ids.TagSet // TODO rewrite as generic interface
		GetTagSetLike() ids.TagSetLike
		GetTagCollection() ids.TagCollection
		AllTags() interfaces.Seq[ids.Tag]
		GetTai() ids.Tai
		GetType() ids.Type
		GetTypeLock() TypeLock

		GetBlobDigest() interfaces.MarklId
		GetObjectDigest() interfaces.MarklId
		GetMotherObjectSig() interfaces.MarklId
		GetRepoPubKey() interfaces.MarklId
		GetObjectSig() interfaces.MarklId
	}

	IMetadataMutable interface {
		interfaces.CommandComponentWriter
		IMetadata
		GetterMutable

		Subtract(IMetadata)

		// TODO rewrite
		AddTagPtr(e *ids.Tag) (err error)
		AddTag(ids.Tag) (err error)
		ResetTags()
		SetTags(ids.TagSet)
		AddTagString(tagString string) (err error)
		AddTagPtrFast(tag *ids.Tag) (err error)
		GenerateExpandedTags()

		GetIndexMutable() IIndexMutable

		GetBlobDigestMutable() interfaces.MutableMarklId
		GetDescriptionMutable() *descriptions.Description
		GetMotherObjectSigMutable() interfaces.MutableMarklId
		GetObjectDigestMutable() interfaces.MutableMarklId
		GetObjectSigMutable() interfaces.MutableMarklId
		GetRepoPubKeyMutable() interfaces.MutableMarklId
		GetTaiMutable() *ids.Tai
		GetTypeMutable() *ids.Type
		GetTypeLockMutable() *TypeLock
	}

	Getter interface {
		GetMetadata() IMetadata
	}

	GetterMutable interface {
		GetMetadataMutable() IMetadataMutable
	}

	PersistentFormatterContext interface {
		Getter
		GetterMutable
	}

	PersistentParserContext interface {
		GetterMutable
	}
)
