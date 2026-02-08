package queries

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/hotel/env_dir"
)

type ErrBlobMissing struct {
	ObjectId
	env_dir.ErrBlobMissing
}

// TODO add recovery text
func (err ErrBlobMissing) Error() string {
	return fmt.Sprintf(
		"Blob for %q with sha %q does not exist locally.",
		err.ObjectId,
		err.BlobId,
	)
}

func (err ErrBlobMissing) Is(target error) bool {
	_, ok := target.(ErrBlobMissing)
	return ok
}

func IsErrBlobMissing(err error) bool {
	return errors.Is(err, ErrBlobMissing{})
}
