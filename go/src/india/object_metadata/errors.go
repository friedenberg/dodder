package object_metadata

import (
	"fmt"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/golf/fd"
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
	inlineBlobDigest interfaces.MarklId,
) (err *ErrHasInlineBlobAndFilePath) {
	err = &ErrHasInlineBlobAndFilePath{}
	err.BlobFD.ResetWith(blobFD)
	err.InlineDigest.SetDigest(inlineBlobDigest)
	return err
}

type ErrHasInlineBlobAndFilePath struct {
	BlobFD       fd.FD
	InlineDigest markl.Id
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
	inline, metadata interfaces.MarklId,
) (err *ErrHasInlineBlobAndMetadataDigest) {
	err = &ErrHasInlineBlobAndMetadataDigest{}
	err.metadata = markl.Clone(metadata)
	err.Inline = markl.Clone(inline)
	return err
}

type ErrHasInlineBlobAndMetadataDigest struct {
	Inline   interfaces.MarklId
	metadata interfaces.MarklId
}

func (err *ErrHasInlineBlobAndMetadataDigest) Error() string {
	return fmt.Sprintf(
		"text has inline blob and metadata blob id: \ninline blob id: %s\n metadata blob id: %s",
		err.Inline,
		err.metadata,
	)
}
