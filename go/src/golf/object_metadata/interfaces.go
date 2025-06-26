package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	Getter interface {
		GetMetadata() *Metadata
	}

	Setter interface {
		SetMetadata(*Metadata)
	}

	MetadataLike interface {
		Getter
	}

	PersistentFormatterContext interface {
		Getter
	}

	PersistentParserContext interface {
		Getter
	}

	TextFormatterContext struct {
		TextFormatterOptions
		PersistentFormatterContext
	}

	TextParserContext interface {
		PersistentParserContext
		SetBlobSha(interfaces.Sha) error
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
