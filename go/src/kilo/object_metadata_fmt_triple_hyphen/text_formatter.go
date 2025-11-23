package object_metadata_fmt_triple_hyphen

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/ohio"
)

type textFormatter struct {
	Dependencies
	sequence []interfaces.FuncWriterElementInterface[TextFormatterContext]
}

func MakeTextFormatterMetadataBlobPath(
	common Dependencies,
) textFormatter {
	return textFormatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[TextFormatterContext]{
			common.writeBoundary,
			common.writeCommonMetadataFormat,
			common.writeBlobPath,
			common.writeTypeAndSigIfNecessary,
			common.writeComments,
			common.writeBoundary,
		},
	}
}

// TODO change format to have blob and type separately
func MakeTextFormatterMetadataOnly(
	deps Dependencies,
) textFormatter {
	return textFormatter{
		Dependencies: deps,
		sequence: []interfaces.FuncWriterElementInterface[TextFormatterContext]{
			deps.writeBoundary,
			deps.writeCommonMetadataFormat,
			deps.writeBlobDigest,
			deps.writeTypeAndSigIfNecessary,
			deps.writeComments,
			deps.writeBoundary,
		},
	}
}

func MakeTextFormatterMetadataInlineBlob(
	common Dependencies,
) textFormatter {
	return textFormatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[TextFormatterContext]{
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

func MakeTextFormatterExcludeMetadata(
	common Dependencies,
) textFormatter {
	return textFormatter{
		Dependencies: common,
		sequence: []interfaces.FuncWriterElementInterface[TextFormatterContext]{
			common.writeBlob,
		},
	}
}

func (formatter textFormatter) FormatMetadata(
	writer io.Writer,
	formatterContext TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteSeq(writer, formatterContext, formatter.sequence...)
}
