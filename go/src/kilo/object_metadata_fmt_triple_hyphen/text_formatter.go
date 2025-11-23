package object_metadata_fmt_triple_hyphen

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
)

type formatter struct {
	Dependencies
	sequence []interfaces.FuncWriterElementInterface[FormatterContext]
}

func MakeFormatterMetadataBlobPath(
	common Dependencies,
) formatter {
	return formatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[FormatterContext]{
			common.writeBoundary,
			common.writeCommonMetadataFormat,
			common.writeBlobPath,
			common.writeTypeAndSigIfNecessary,
			common.writeComments,
			common.writeBoundary,
		},
	}
}

func MakeFormatterMetadataOnly(
	deps Dependencies,
) formatter {
	return formatter{
		Dependencies: deps,
		sequence: []interfaces.FuncWriterElementInterface[FormatterContext]{
			deps.writeBoundary,
			deps.writeCommonMetadataFormat,
			deps.writeBlobDigest,
			deps.writeTypeAndSigIfNecessary,
			deps.writeComments,
			deps.writeBoundary,
		},
	}
}

func MakeFormatterMetadataInlineBlob(
	common Dependencies,
) formatter {
	return formatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[FormatterContext]{
			common.writeBoundary,
			common.writeCommonMetadataFormat,
			common.writeTypeAndSigIfNecessary,
			common.writeComments,
			common.writeBoundary,
			common.writeNewLine,
			common.writeBlob,
		},
	}
}

func MakeFormatterExcludeMetadata(
	common Dependencies,
) formatter {
	return formatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[FormatterContext]{
			common.writeBlob,
		},
	}
}

func (formatter formatter) FormatMetadata(
	writer io.Writer,
	formatterContext FormatterContext,
) (n int64, err error) {
	return ohio.WriteSeq(writer, formatterContext, formatter.sequence...)
}
