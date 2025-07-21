package object_metadata

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
)

type textParser struct {
	awf interfaces.BlobWriter
	af  script_config.RemoteScript
}

func MakeTextParser(
	dependencies Dependencies,
) TextParser {
	if dependencies.BlobStore == nil {
		panic("nil BlobWriterFactory")
	}

	return textParser{
		awf: dependencies.BlobStore,
		af:  dependencies.BlobFormatter,
	}
}

func (f textParser) ParseMetadata(
	r io.Reader,
	c TextParserContext,
) (n int64, err error) {
	m := c.GetMetadata()
	Resetter.Reset(m)

	var n1 int64

	defer func() {
		c.SetBlobSha(&m.Blob)
	}()

	mp := &textParser2{
		BlobWriter:        f.awf,
		TextParserContext: c,
	}

	var blobWriter sha.WriteCloser

	if blobWriter, err = f.awf.BlobWriter(); err != nil {
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

	if n, err = mr.ReadFrom(r); err != nil {
		n += n1
		err = errors.Wrap(err)
		return
	}

	n += n1

	inlineBlobSha := sha.Make(blobWriter.GetDigest())

	if !m.Blob.IsNull() && !mp.Blob.GetDigest().IsNull() {
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

		m.Blob.SetShaLike(mp.Blob.GetDigest())
	}

	switch {
	case m.Blob.IsNull() && !inlineBlobSha.IsNull():
		m.Blob.SetShaLike(inlineBlobSha)

	case !m.Blob.IsNull() && inlineBlobSha.IsNull():
		// noop

	case !m.Blob.IsNull() && !inlineBlobSha.IsNull() &&
		!m.Blob.Equals(inlineBlobSha):
		err = errors.Wrap(
			MakeErrHasInlineBlobAndMetadataSha(
				inlineBlobSha,
				&m.Blob,
			),
		)

		return
	}

	return
}
