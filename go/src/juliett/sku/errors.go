package sku

import (
	"errors"
	"fmt"
)

func MakeErrMergeConflict(item *FSItem) (err *ErrMergeConflict) {
	err = &ErrMergeConflict{}

	if item != nil {
		err.ResetWith(item)
	}

	return
}

type ErrMergeConflict struct {
	FSItem
}

func (e *ErrMergeConflict) Is(target error) bool {
	_, ok := target.(*ErrMergeConflict)
	return ok
}

func (e *ErrMergeConflict) Error() string {
	return fmt.Sprintf(
		"merge conflict for fds: Object: %q, Blob: %q",
		&e.Object,
		&e.Blob,
	)
}

func IsErrMergeConflict(err error) bool {
	return errors.Is(err, &ErrMergeConflict{})
}
