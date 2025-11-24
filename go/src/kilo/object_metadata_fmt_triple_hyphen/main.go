package object_metadata_fmt_triple_hyphen

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

type (
	Formatter interface {
		FormatMetadata(io.Writer, FormatterContext) (int64, error)
	}

	Parser interface {
		ParseMetadata(io.Reader, ParserContext) (int64, error)
	}

	FormatterOptions = checkout_options.TextFormatterOptions

	FormatterContext struct {
		FormatterOptions
		object_metadata.PersistentFormatterContext
	}

	ParserContext interface {
		object_metadata.PersistentParserContext
		SetBlobDigest(interfaces.MarklId) error
	}

	FormatterFamily struct {
		BlobPath     Formatter
		InlineBlob   Formatter
		MetadataOnly Formatter
		BlobOnly     Formatter
	}

	Format struct {
		FormatterFamily
		Parser
	}

	funcWrite = interfaces.FuncWriterElementInterface[FormatterContext]
)
