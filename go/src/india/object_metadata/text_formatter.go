package object_metadata

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
			common.writePathType,
			common.writeComments,
			common.writeBoundary,
		},
	}
}

func MakeTextFormatterMetadataOnly(
	deps Dependencies,
) textFormatter {
	return textFormatter{
		Dependencies: deps,
		sequence: []interfaces.FuncWriterElementInterface[TextFormatterContext]{
			deps.writeBoundary,
			deps.writeCommonMetadataFormat,
			deps.writeBlobDigestAndType,
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
			common.writeTyp,
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

func (f textFormatter) FormatMetadata(
	w io.Writer,
	c TextFormatterContext,
) (n int64, err error) {
	return ohio.WriteSeq(w, c, f.sequence...)
}
