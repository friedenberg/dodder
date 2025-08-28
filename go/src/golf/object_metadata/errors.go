package object_metadata

import (
	"fmt"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

func MakeErrBlobFormatterFailed(
	err *exec.ExitError,
) ErrBlobFormatterFailed {
	return ErrBlobFormatterFailed{ExitError: err}
}

type ErrBlobFormatterFailed struct {
	*exec.ExitError
}

func (err ErrBlobFormatterFailed) Error() string {
	return fmt.Sprintf(
		"blob formatter failed (exit status: %d): %q",
		err.ExitCode(),
		err.Stderr,
	)
}

func (err ErrBlobFormatterFailed) Is(target error) bool {
	_, ok := target.(ErrBlobFormatterFailed)
	return ok
}

func (err ErrBlobFormatterFailed) ShouldShowStackTrace() bool {
	return false
}

func MakeErrHasInlineBlobAndFilePath(
	blobFD *fd.FD,
	inlineBlobDigest interfaces.BlobId,
) (err *ErrHasInlineBlobAndFilePath) {
	err = &ErrHasInlineBlobAndFilePath{}
	err.BlobFD.ResetWith(blobFD)
	err.InlineDigest.SetDigest(inlineBlobDigest)
	return
}

type ErrHasInlineBlobAndFilePath struct {
	BlobFD       fd.FD
	InlineDigest merkle.Id
}

func (err *ErrHasInlineBlobAndFilePath) Error() string {
	return fmt.Sprintf(
		"text has inline blob and file: \nexternal path: %s\nexternal sha: %s\ninline sha: %s",
		err.BlobFD.GetPath(),
		err.BlobFD.GetDigest(),
		&err.InlineDigest,
	)
}

func MakeErrHasInlineBlobAndMetadataBlobId(
	inline, metadata interfaces.BlobId,
) (err *ErrHasInlineBlobAndMetadataSha) {
	err = &ErrHasInlineBlobAndMetadataSha{}
	err.Metadata = merkle_ids.Clone(metadata)
	err.Inline = merkle_ids.Clone(inline)
	return
}

type ErrHasInlineBlobAndMetadataSha struct {
	Inline   interfaces.BlobId
	Metadata interfaces.BlobId
}

func (err *ErrHasInlineBlobAndMetadataSha) Error() string {
	return fmt.Sprintf(
		"text has inline blob and metadata blob id: \ninline blob id: %s\n metadata blob id: %s",
		err.Inline,
		err.Metadata,
	)
}
