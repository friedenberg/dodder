package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type textParser struct {
	hashType      interfaces.FormatHash
	blobWriter    interfaces.BlobWriterFactory
	blobFormatter script_config.RemoteScript
}

func MakeTextParser(
	dependencies Dependencies,
) TextParser {
	if dependencies.BlobStore == nil {
		panic("nil BlobWriterFactory")
	}

	return textParser{
		hashType:      dependencies.GetBlobDigestType(),
		blobWriter:    dependencies.BlobStore,
		blobFormatter: dependencies.BlobFormatter,
	}
}

func (parser textParser) ParseMetadata(
	reader io.Reader,
	context TextParserContext,
) (n int64, err error) {
	metadata := context.GetMetadata()
	Resetter.Reset(metadata)

	var n1 int64

	parser2 := &textParser2{
		BlobWriterFactory: parser.blobWriter,
		hashType:          parser.hashType,
		TextParserContext: context,
	}

	var blobWriter interfaces.BlobWriter

	if blobWriter, err = parser.blobWriter.MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	if blobWriter == nil {
		err = errors.ErrorWithStackf("blob writer is nil")
		return
	}

	defer errors.DeferredCloser(&err, blobWriter)

	metadataReader := triple_hyphen_io.Reader{
		Metadata: parser2,
		Blob:     blobWriter,
	}

	if n, err = metadataReader.ReadFrom(reader); err != nil {
		n += n1
		err = errors.Wrap(err)
		return
	}

	n += n1

	inlineBlobDigest := blobWriter.GetMarklId()

	if !metadata.DigBlob.IsNull() && !parser2.Blob.GetDigest().IsNull() {
		err = errors.Wrap(
			MakeErrHasInlineBlobAndFilePath(
				&parser2.Blob,
				inlineBlobDigest,
			),
		)

		return
	} else if !parser2.Blob.GetDigest().IsNull() {
		metadata.Fields = append(
			metadata.Fields,
			Field{
				Key:       "blob",
				Value:     parser2.Blob.GetPath(),
				ColorType: string_format_writer.ColorTypeId,
			},
		)

		metadata.DigBlob.SetDigest(parser2.Blob.GetDigest())
	}

	switch {
	case metadata.DigBlob.IsNull() && !inlineBlobDigest.IsNull():
		metadata.DigBlob.SetDigest(inlineBlobDigest)

	case !metadata.DigBlob.IsNull() && inlineBlobDigest.IsNull():
		// noop

	case !metadata.DigBlob.IsNull() && !inlineBlobDigest.IsNull() &&
		!markl.Equals(&metadata.DigBlob, inlineBlobDigest):
		err = errors.Wrap(
			MakeErrHasInlineBlobAndMetadataBlobId(
				inlineBlobDigest,
				&metadata.DigBlob,
			),
		)

		return
	}

	return
}
