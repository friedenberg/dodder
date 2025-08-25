package env_dir

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
)

func IsErrBlobAlreadyExists(err error) bool {
	return errors.Is(err, ErrBlobAlreadyExists{})
}

func MakeErrBlobAlreadyExists(
	blobId interfaces.BlobId,
	path string,
) ErrBlobAlreadyExists {
	return ErrBlobAlreadyExists{
		Path:   path,
		BlobId: merkle_ids.Clone(blobId),
	}
}

type ErrBlobAlreadyExists struct {
	BlobId interfaces.BlobId
	Path   string
}

func (err ErrBlobAlreadyExists) Error() string {
	return fmt.Sprintf(
		"File with blob_id %s already exists: %s",
		merkle_ids.Format(err.BlobId),
		err.Path,
	)
}

func (err ErrBlobAlreadyExists) Is(target error) bool {
	_, ok := target.(ErrBlobAlreadyExists)
	return ok
}

func IsErrBlobMissing(err error) bool {
	return errors.Is(err, ErrBlobMissing{})
}

type ErrBlobMissing struct {
	interfaces.BlobIdGetter
	Path string
}

func (e ErrBlobMissing) Error() string {
	if e.Path == "" {
		return fmt.Sprintf(
			"Blob with sha %q does not exist locally",
			e.GetBlobId(),
		)
	} else {
		return fmt.Sprintf(
			"Blob with sha %q does not exist locally: %q",
			e.GetBlobId(),
			e.Path,
		)
	}
}

func (e ErrBlobMissing) Is(target error) bool {
	_, ok := target.(ErrBlobMissing)
	return ok
}

func MakeErrTempAlreadyExists(
	path string,
) (err ErrTempAlreadyExists) {
	err = ErrTempAlreadyExists{Path: path}
	return
}

var _ interfaces.ErrorHelpful = ErrTempAlreadyExists{}

type ErrTempAlreadyExists struct {
	Path string
}

func (err ErrTempAlreadyExists) Error() string {
	return fmt.Sprintf("Local temporary directory already exists: %q", err.Path)
}

func (err ErrTempAlreadyExists) GetErrorCause() []string {
	return []string{
		"Another dodder previous process with the same PID likely terminated unexpectedly",
	}
}

func (err ErrTempAlreadyExists) GetErrorRecovery() []string {
	return []string{
		"Check if there are any relevant files in the directory, or possible delete it",
	}
}

func (err ErrTempAlreadyExists) Is(target error) bool {
	_, ok := target.(ErrTempAlreadyExists)
	return ok
}
