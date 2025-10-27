package blob_stores

import (
	"io"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

func CopyBlobIfNecessary(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	expectedDigest interfaces.MarklId,
	extraWriter io.Writer,
	hashType interfaces.FormatHash,
) (copyResult CopyResult) {
	copyResult.BlobId = expectedDigest

	// first check if we have the blob already, intentionally before checking if
	// `src` is non-nil, as this method is also used to emit a list of missing
	// blobs in the remote transfer protocol
	if dst.HasBlob(expectedDigest) {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateExistsLocally
		return copyResult
	}

	if src == nil {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateNilRemoteBlobStore
		return copyResult
	}

	if err := markl.AssertIdIsNotNull(expectedDigest); err != nil {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateExistsLocallyAndRemotely
		return copyResult
	}

	errors.PanicIfError(markl.AssertIdIsNotNull(expectedDigest))

	var readCloser interfaces.BlobReader

	{
		var err error

		if readCloser, err = src.MakeBlobReader(expectedDigest); err != nil {
			copyResult.SetError(err)
			return copyResult
		}
	}

	defer errors.ContextMustClose(env, readCloser)

	var writeCloser interfaces.BlobWriter

	if hashType == nil {
		var err error

		if hashType, err = markl.GetFormatHashOrError(
			expectedDigest.GetMarklFormat().GetMarklFormatId(),
		); err != nil {
			copyResult.SetError(err)
			return copyResult
		}
	}

	{
		var err error

		if writeCloser, err = dst.MakeBlobWriter(hashType); err != nil {
			copyResult.SetError(err)
			return copyResult
		}
	}

	defer errors.ContextMustClose(env, writeCloser)

	outputWriter := io.Writer(writeCloser)

	if extraWriter != nil {
		outputWriter = io.MultiWriter(outputWriter, extraWriter)
	}

	{
		var err error

		copyResult.bytesWritten, err = io.Copy(
			outputWriter,
			readCloser,
		)
		if err != nil {
			copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
			return copyResult
		} else {
			copyResult.state = CopyResultStateSuccess
		}
	}

	readerDigest := readCloser.GetMarklId()

	writerDigest := writeCloser.GetMarklId()

	if !markl.Equals(readerDigest, expectedDigest) {
		copyResult.setErrorAfterCopy(
			copyResult.bytesWritten,
			errors.Errorf(
				"lookup digest was %s while read digest was %s",
				expectedDigest,
				readerDigest,
			),
		)

		return copyResult
	}

	if err := markl.AssertEqual(expectedDigest, writerDigest); err != nil {
		copyResult.setErrorAfterCopy(
			copyResult.bytesWritten,
			err,
		)

		return copyResult
	}

	return copyResult
}

func CopyReaderToWriter(
	ctx interfaces.Context,
	dst interfaces.BlobWriter,
	src io.Reader,
	expected interfaces.MarklId,
	extraWriter io.Writer,
	heartbeats func(time time.Time),
	pulse time.Duration,
) (copyResult CopyResult) {
	var writer io.Writer = dst

	if extraWriter != nil {
		writer = io.MultiWriter(dst, extraWriter)
	}

	if err := errors.RunChildContextWithPrintTicker(
		ctx,
		func(ctx interfaces.Context) {
			var err error

			if copyResult.bytesWritten, err = io.Copy(writer, src); err != nil {
				ctx.Cancel(err)
			}
		},
		heartbeats,
		pulse,
	); err != nil {
		copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
		return copyResult
	} else {
		copyResult.state = CopyResultStateSuccess
	}

	if err := dst.Close(); err != nil {
		copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
		return copyResult
	}

	copyResult.BlobId = dst.GetMarklId()

	if expected != nil {
		if err := markl.AssertEqual(expected, copyResult.BlobId); err != nil {
			copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
			return copyResult
		}
	}

	return copyResult
}
