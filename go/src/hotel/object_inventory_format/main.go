package object_inventory_format

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
)

type (
	FormatterContext interface {
		object_metadata.PersistentFormatterContext
		GetObjectId() *ids.ObjectId
	}

	ParserContext interface {
		object_metadata.PersistentParserContext
		SetObjectIdLike(interfaces.ObjectId) error
	}

	Formatter interface {
		FormatPersistentMetadata(
			io.Writer,
			FormatterContext,
			Options,
		) (int64, error)
	}

	Parser interface {
		ParsePersistentMetadata(
			*catgut.RingBuffer,
			ParserContext,
			Options,
		) (int64, error)
	}

	Format interface {
		Formatter
		Parser
	}
)

func Default() Format {
	return v4{}
}

func FormatForVersion(sv interfaces.StoreVersion) Format {
	if store_version.Equals(
		sv,
		store_version.V3,
		store_version.V4,
	) {
		return v4{}
	} else {
		return v5{}
	}
}

func FormatForVersions(write, read interfaces.StoreVersion) Format {
	return MakeBespoke(
		FormatForVersion(write),
		FormatForVersion(read),
	)
}

type nopFormatterContext struct {
	object_metadata.PersistentFormatterContext
}

func (nopFormatterContext) GetObjectId() *ids.ObjectId {
	return nil
}
