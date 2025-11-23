package object_metadata_fmt_triple_hyphen

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
)

type (
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
	TextFormatterContext struct {
		TextFormatterOptions
		object_metadata.PersistentFormatterContext
	}

	TextParserContext interface {
		object_metadata.PersistentParserContext
		SetBlobDigest(interfaces.MarklId) error
	}
)
