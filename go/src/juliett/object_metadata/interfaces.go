package object_metadata

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
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

		GetDescription() descriptions.Description
		GetIndex() IIndex
		GetTags() ids.TagSet
		GetTai() ids.Tai
		GetType() ids.Type
		GetTypeTuple() TypeTuple

		GetBlobDigest() interfaces.MarklId
		GetObjectDigest() interfaces.MarklId
		GetMotherObjectSig() interfaces.MarklId
		GetRepoPubKey() interfaces.MarklId
		GetObjectSig() interfaces.MarklId
		GetSelfWithoutTai() interfaces.MarklId
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

		GetBlobDigestMutable() interfaces.MutableMarklId
		GetDescriptionMutable() *descriptions.Description
		GetMotherObjectSigMutable() interfaces.MutableMarklId
		GetObjectDigestMutable() interfaces.MutableMarklId
		GetObjectSigMutable() interfaces.MutableMarklId
		GetRepoPubKeyMutable() interfaces.MutableMarklId
		GetSelfWithoutTaiMutable() interfaces.MutableMarklId
		GetTaiMutable() *ids.Tai
		GetTypeMutable() *ids.Type
		GetTypeTupleMutable() *TypeTuple
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
