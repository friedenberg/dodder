package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	// Metadata = Metadata

	Getter interface {
		GetMetadata() Metadata
	}

	GetterMutable interface {
		GetMetadataMutable() *Metadata
	}

	Setter interface {
		SetMetadata(*Metadata)
	}

	MetadataLike interface {
		GetterMutable
	}

	PersistentFormatterContext interface {
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
