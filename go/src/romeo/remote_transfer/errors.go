package remote_transfer

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

var (
	ErrSkipped = errors.New("skipped due to exclude objects option")

	ErrNeedsMerge = errors.Err409Conflict.Errorf(
		"import failed with conflicts, merging required",
	)
)
