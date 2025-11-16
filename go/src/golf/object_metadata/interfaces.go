package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/collections_slice"
	"code.linenisgreat.com/dodder/go/src/echo/descriptions"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type (
	Metadata        = metadata
	MetadataMutable = *metadata

	IMetadata interface {
		Getter

		GetTags() ids.TagSet
		GetIndex() IIndex
		GetLockfile() Lockfile
		GetDescription() descriptions.Description
		GetTai() ids.Tai
		GetType() ids.Type
		GetComments() interfaces.Seq[string]
		GetFields() interfaces.Seq[Field]

		GetBlobDigest() interfaces.MarklId
		GetObjectDigest() interfaces.MarklId
		GetMotherObjectSig() interfaces.MarklId
		GetRepoPubKey() interfaces.MarklId
		GetObjectSig() interfaces.MarklId
		GetSelfWithoutTai() interfaces.MarklId
	}

	IMetadataMutable interface {
		IMetadata
		GetterMutable

		AddTagPtr(e *ids.Tag) (err error)
		ResetTags()
		SetTags(ids.TagSet)
		SetTagsFast(ids.TagSet)
		AddTagString(tagString string) (err error)
		AddTagPtrFast(tag *ids.Tag) (err error)
		GenerateExpandedTags()

		GetCommentsMutable() *collections_slice.Slice[string]
		GetFieldsMutable() *collections_slice.Slice[Field]
		GetIndexMutable() IIndexMutable
		GetLockfileMutable() LockfileMutable
		// TODO rename to GetTypeMutable
		GetTypePtr() *ids.Type
		GetDescriptionMutable() *descriptions.Description
		GetTaiMutable() *ids.Tai
		GetBlobDigestMutable() interfaces.MutableMarklId
		GetObjectDigestMutable() interfaces.MutableMarklId
		GetMotherObjectSigMutable() interfaces.MutableMarklId
		GetRepoPubKeyMutable() interfaces.MutableMarklId
		GetObjectSigMutable() interfaces.MutableMarklId
		GetSelfWithoutTaiMutable() interfaces.MutableMarklId
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

	TextFormatterContext struct {
		TextFormatterOptions
		PersistentFormatterContext
	}

	TextParserContext interface {
		PersistentParserContext
		SetBlobDigest(interfaces.MarklId) error
	}

	TextFormatOutput struct {
		io.Writer
		string
	}

	TextFormatter interface {
		FormatMetadata(io.Writer, TextFormatterContext) (int64, error)
	}

	TextParser interface {
		ParseMetadata(io.Reader, TextParserContext) (int64, error)
	}
)
