package sku_fmt

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type itemDeletedStringFormatWriter struct {
	interfaces.Config
	rightAlignedWriter   interfaces.StringEncoderTo[string]
	idStringFormatWriter interfaces.StringEncoderTo[string]
	fieldsFormatWriter   interfaces.StringEncoderTo[string_format_writer.Box]
}

func MakeItemDeletedStringWriterFormat(
	config interfaces.Config,
	co string_format_writer.ColorOptions,
	fieldsFormatWriter interfaces.StringEncoderTo[string_format_writer.Box],
) *itemDeletedStringFormatWriter {
	return &itemDeletedStringFormatWriter{
		Config:             config,
		rightAlignedWriter: string_format_writer.MakeRightAligned(),
		idStringFormatWriter: string_format_writer.MakeColor(
			co,
			string_format_writer.MakeString[string](),
			string_format_writer.ColorTypeId,
		),
		fieldsFormatWriter: fieldsFormatWriter,
	}
}

func (f *itemDeletedStringFormatWriter) EncodeStringTo(
	o *sku.Transacted,
	sw interfaces.WriterAndStringWriter,
) (n int64, err error) {
	var (
		n1 int
		n2 int64
	)

	prefixOne := string_format_writer.StringDeleted

	if f.IsDryRun() {
		prefixOne = string_format_writer.StringWouldDelete
	}

	n2, err = f.rightAlignedWriter.EncodeStringTo(prefixOne, sw)
	n += n2

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = sw.WriteString("[")
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n2, err = f.fieldsFormatWriter.EncodeStringTo(
		string_format_writer.Box{Contents: o.Metadata.Fields},
		sw,
	)
	n += n2

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = sw.WriteString("]")
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
