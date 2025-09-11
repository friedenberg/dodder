package blob_stores

import (
	"fmt"
	"io"
	"strings"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

type CopyResultState interface {
	state()
}

//go:generate stringer -type=copyResultState
type copyResultState int

func (copyResultState) state() {}

const (
	CopyResultStateUnknown = copyResultState(iota)
	CopyResultStateSuccess
	CopyResultStateMissingLocally
	CopyResultStateExistsLocally
	CopyResultStateExistsLocallyAndRemotely
	CopyResultStateError
)

type CopyResult struct {
	BlobId       interfaces.MarklId // may not be nil
	bytesWritten int64
	state        copyResultState
	error
}

func (copyResult CopyResult) String() string {
	var stringsBuilder strings.Builder

	stringsBuilder.WriteString("CopyResult{")

	if copyResult.BlobId == nil {
		stringsBuilder.WriteString("BlobId: nil, ")
	} else {
		fmt.Fprintf(&stringsBuilder, "BlobId: %s, ", copyResult.BlobId.StringWithFormat())
	}

	fmt.Fprintf(&stringsBuilder, "BytesWritten: %d, ", copyResult.bytesWritten)
	fmt.Fprintf(&stringsBuilder, "State: %s, ", copyResult.state)

	if copyResult.error == nil {
		stringsBuilder.WriteString("Error: nil")
	} else {
		fmt.Fprintf(&stringsBuilder, "Error: %d", copyResult.error)
	}

	stringsBuilder.WriteString("}")

	return stringsBuilder.String()
}

func (copyResult *CopyResult) setErrorAfterCopy(n int64, err error) {
	copyResult.bytesWritten = n

	if env_dir.IsErrBlobAlreadyExists(err) {
		copyResult.SetBlobExistsLocallyAndRemotely()
	} else if markl.IsErrNull(err) {
		copyResult.SetBlobExistsLocallyAndRemotely()
	} else if env_dir.IsErrBlobMissing(err) {
		copyResult.SetBlobMissingLocally()
	} else if err != nil {
		copyResult.state = CopyResultStateError
		copyResult.error = err
	} else {
		copyResult.state = CopyResultStateSuccess
	}
}

func (copyResult CopyResult) GetBytesWrittenAndState() (int64, CopyResultState) {
	return copyResult.bytesWritten, copyResult.state
}

func (copyResult CopyResult) IsError() bool {
	return copyResult.state == CopyResultStateError
}

func (copyResult CopyResult) GetError() error {
	return copyResult.error
}

func (copyResult CopyResult) IsMissing() bool {
	return copyResult.state == CopyResultStateMissingLocally
}

func (copyResult CopyResult) Exists() bool {
	switch copyResult.state {
	case CopyResultStateExistsLocally,
		CopyResultStateExistsLocallyAndRemotely:
		return true

	default:
		return false
	}
}

func (copyResult *CopyResult) SetBlobMissingLocally() {
	copyResult.bytesWritten = -1
	copyResult.state = CopyResultStateMissingLocally
}

func (copyResult *CopyResult) SetBlobExistsLocally() {
	copyResult.bytesWritten = -1
	copyResult.state = CopyResultStateExistsLocally
}

func (copyResult *CopyResult) SetBlobExistsLocallyAndRemotely() {
	copyResult.bytesWritten = -1
	copyResult.state = CopyResultStateExistsLocallyAndRemotely
}

// TODO offer option to decide which hash type
func CopyBlobIfNecessary(
	env env_ui.Env,
	dst interfaces.BlobStore,
	src interfaces.BlobStore,
	expectedDigest interfaces.MarklId,
	extraWriter io.Writer,
) (copyResult CopyResult) {
	copyResult.BlobId = expectedDigest

	// first check if we have the blob already, intentionally before checking if
	// `src` is non-nil, as this method is also used to emit a list of missing
	// blobs in the remote transfer protocol
	if dst.HasBlob(expectedDigest) {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateExistsLocally
		return
	}

	if src == nil {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateMissingLocally
		return
	}

	if err := markl.AssertIdIsNotNull(expectedDigest, ""); err != nil {
		copyResult.bytesWritten = -1
		copyResult.state = CopyResultStateExistsLocallyAndRemotely
		return
	}

	errors.PanicIfError(markl.AssertIdIsNotNull(expectedDigest, ""))

	var readCloser interfaces.BlobReader

	{
		var err error

		if readCloser, err = src.MakeBlobReader(expectedDigest); err != nil {
			copyResult.setErrorAfterCopy(0, err)
			return
		}
	}

	defer errors.ContextMustClose(env, readCloser)

	var writeCloser interfaces.BlobWriter

	var hashType markl.HashType

	{
		var err error

		if hashType, err = markl.GetHashTypeOrError(
			expectedDigest.GetMarklType().GetMarklTypeId(),
		); err != nil {
			copyResult.setErrorAfterCopy(0, err)
			return
		}
	}

	{
		var err error

		if writeCloser, err = dst.MakeBlobWriter(hashType); err != nil {
			copyResult.setErrorAfterCopy(0, err)
			return
		}
	}

	// TODO should this be closed with an error when the digests don't match to
	// prevent a garbage object in the store?
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
			return
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

		return
	}

	if err := markl.AssertEqual(expectedDigest, writerDigest); err != nil {
		copyResult.setErrorAfterCopy(
			copyResult.bytesWritten,
			err,
		)

		return
	}

	return
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
		return
	} else {
		copyResult.state = CopyResultStateSuccess
	}

	if err := dst.Close(); err != nil {
		copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
		return
	}

	copyResult.BlobId = dst.GetMarklId()

	if expected != nil {
		if err := markl.AssertEqual(expected, copyResult.BlobId); err != nil {
			copyResult.setErrorAfterCopy(copyResult.bytesWritten, err)
			return
		}
	}

	return
}
