package env_dir

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

func MakeErrAlreadyExists(
	sh interfaces.Sha,
	path string,
) (err *ErrAlreadyExists) {
	err = &ErrAlreadyExists{Path: path}
	err.Sha.SetShaLike(sh)
	return
}

type ErrAlreadyExists struct {
	sha.Sha
	Path string
}

func (e *ErrAlreadyExists) Error() string {
	return fmt.Sprintf("File with sha %s already exists: %s", &e.Sha, e.Path)
}

func (e *ErrAlreadyExists) Is(target error) bool {
	_, ok := target.(*ErrAlreadyExists)
	return ok
}

func IsErrBlobMissing(err error) bool {
	return errors.Is(err, ErrBlobMissing{})
}

type ErrBlobMissing struct {
	interfaces.DigestGetter
	Path string
}

func (e ErrBlobMissing) Error() string {
	if e.Path == "" {
		return fmt.Sprintf(
			"Blob with sha %q does not exist locally",
			e.GetDigest(),
		)
	} else {
		return fmt.Sprintf(
			"Blob with sha %q does not exist locally: %q",
			e.GetDigest(),
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
) (err *ErrTempAlreadyExists) {
	err = &ErrTempAlreadyExists{Path: path}
	return
}

type ErrTempAlreadyExists struct {
	Path string
}

func (e *ErrTempAlreadyExists) Error() string {
	return fmt.Sprintf("Local temporary directory already exists: %q", e.Path)
}

func (e *ErrTempAlreadyExists) ErrorCause() string {
	return "Another dodder previous process with the same PID likely terminated unexpectedly"
}

func (e *ErrTempAlreadyExists) ErrorRecovery() string {
	return "Check if there are any relevant files in the directory, or possible delete it"
}

func (e *ErrTempAlreadyExists) ErrorRecoveryAutomatic() string {
	return "TODO"
}

func (e *ErrTempAlreadyExists) Is(target error) bool {
	_, ok := target.(*ErrTempAlreadyExists)
	return ok
}
