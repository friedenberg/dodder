package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type textParser struct {
	blobDigestType string
	blobWriter     interfaces.BlobWriter
	blobFormatter  script_config.RemoteScript
}

func MakeTextParser(
	dependencies Dependencies,
) TextParser {
	if dependencies.BlobStore == nil {
		panic("nil BlobWriterFactory")
	}

	return textParser{
		blobDigestType: dependencies.GetBlobDigestType(),
		blobWriter:     dependencies.BlobStore,
		blobFormatter:  dependencies.BlobFormatter,
	}
}

func (parser textParser) ParseMetadata(
	reader io.Reader,
	context TextParserContext,
) (n int64, err error) {
	m := context.GetMetadata()
	Resetter.Reset(m)

	var n1 int64

	defer func() {
		context.SetBlobDigest(&m.DigBlob)
	}()

	mp := &textParser2{
		BlobWriter:        parser.blobWriter,
		blobDigestType:    parser.blobDigestType,
		TextParserContext: context,
	}

	var blobWriter interfaces.WriteCloseBlobIdGetter

	if blobWriter, err = parser.blobWriter.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if blobWriter == nil {
		err = errors.ErrorWithStackf("blob writer is nil")
		return
	}

	defer errors.DeferredCloser(&err, blobWriter)

	mr := triple_hyphen_io.Reader{
		Metadata: mp,
		Blob:     blobWriter,
	}

	if n, err = mr.ReadFrom(reader); err != nil {
		n += n1
		err = errors.Wrap(err)
		return
	}

	n += n1

	inlineBlobSha := sha.MustWithDigester(blobWriter)

	if !m.DigBlob.IsNull() && !mp.Blob.GetDigest().IsNull() {
		err = errors.Wrap(
			MakeErrHasInlineBlobAndFilePath(
				&mp.Blob,
				inlineBlobSha,
			),
		)

		return
	} else if !mp.Blob.GetDigest().IsNull() {
		m.Fields = append(
			m.Fields,
			Field{
				Key:       "blob",
				Value:     mp.Blob.GetPath(),
				ColorType: string_format_writer.ColorTypeId,
			},
		)

		m.DigBlob.SetDigest(mp.Blob.GetDigest())
	}

	switch {
	case m.DigBlob.IsNull() && !inlineBlobSha.IsNull():
		m.DigBlob.SetDigest(inlineBlobSha)

	case !m.DigBlob.IsNull() && inlineBlobSha.IsNull():
		// noop

	case !m.DigBlob.IsNull() && !inlineBlobSha.IsNull() &&
		!merkle_ids.Equals(&m.DigBlob, inlineBlobSha):
		err = errors.Wrap(
			MakeErrHasInlineBlobAndMetadataBlobId(
				inlineBlobSha,
				&m.DigBlob,
			),
		)

		return
	}

	return
}
