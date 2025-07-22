package digests

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var ErrIsNull = errors.New("digest is null")

func MakeErrIsNull(digest interfaces.Digest) error {
	if digest.IsNull() {
		return errors.WrapSkip(1, ErrIsNull)
	}

	return nil
}

type ErrNotEqual struct {
	Expected, Actual interfaces.Digest
}

func MakeErrNotEqual(expected, actual interfaces.Digest) *ErrNotEqual {
	err := &ErrNotEqual{
		Expected: expected,
		Actual:   actual,
	}

	return err
}

func (e *ErrNotEqual) Error() string {
	return fmt.Sprintf("expected digest %s but got %s", e.Expected, e.Actual)
}

func (e *ErrNotEqual) Is(target error) bool {
	_, ok := target.(*ErrNotEqual)
	return ok
}

type errLength [2]int

// TODO add another "wrong hasher" error type
func MakeErrLength(expected, actual int) error {
	if expected != actual {
		return errLength{expected, actual}
	} else {
		return nil
	}
}

func (e errLength) Error() string {
	return fmt.Sprintf(
		"expected digest to have length %d, but got %d",
		e[0],
		e[1],
	)
}
