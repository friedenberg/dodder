package object_metadata_fmt_triple_hyphen

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

type (
	Formatter interface {
		FormatMetadata(io.Writer, FormatterContext) (int64, error)
	}

	Parser interface {
		ParseMetadata(io.Reader, ParserContext) (int64, error)
	}

	FormatterContext struct {
		FormatterOptions
		object_metadata.PersistentFormatterContext
	}

	ParserContext interface {
		object_metadata.PersistentParserContext
		SetBlobDigest(interfaces.MarklId) error
	}
)
