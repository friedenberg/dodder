package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
	"code.linenisgreat.com/dodder/go/src/foxtrot/descriptions"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
)

type (
	Metadata        = metadata
	MetadataMutable = *metadata

	TypeTuple = markl.KeyValueTuple[ids.Type, *ids.Type]

	IMetadata interface {
		Getter

		GetTags() ids.TagSet
		GetIndex() IIndex
		GetLockfile() Lockfile
		GetDescription() descriptions.Description
		GetTai() ids.Tai
		GetType() ids.Type
		GetTypeTuple() TypeTuple

		GetBlobDigest() interfaces.MarklId
		GetObjectDigest() interfaces.MarklId
		GetMotherObjectSig() interfaces.MarklId
		GetRepoPubKey() interfaces.MarklId
		GetObjectSig() interfaces.MarklId
		GetSelfWithoutTai() interfaces.MarklId

		GetComments() interfaces.Seq[string] // TODO move to IIndex
	}

	IMetadataMutable interface {
		interfaces.CommandComponentWriter
		IMetadata
		GetterMutable

		AddTagPtr(e *ids.Tag) (err error)
		ResetTags()
		SetTags(ids.TagSet)
		SetTagsFast(ids.TagSet)
		AddTagString(tagString string) (err error)
		AddTagPtrFast(tag *ids.Tag) (err error)
		GenerateExpandedTags()

		GetIndexMutable() IIndexMutable
		GetLockfileMutable() LockfileMutable
		GetTypeMutable() *ids.Type
		GetTypeTupleMutable() *TypeTuple
		GetDescriptionMutable() *descriptions.Description
		GetTaiMutable() *ids.Tai
		GetBlobDigestMutable() interfaces.MutableMarklId
		GetObjectDigestMutable() interfaces.MutableMarklId
		GetMotherObjectSigMutable() interfaces.MutableMarklId
		GetRepoPubKeyMutable() interfaces.MutableMarklId
		GetObjectSigMutable() interfaces.MutableMarklId
		GetSelfWithoutTaiMutable() interfaces.MutableMarklId

		GetCommentsMutable() *collections_slice.Slice[string] // TODO move to IIndexMutable
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
