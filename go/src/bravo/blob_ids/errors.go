package blob_ids

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var ErrIsNull = errors.New("digest is null")

func MakeErrIsNull(binaryId interfaces.BinaryId) error {
	if binaryId.IsNull() {
		return errors.WrapSkip(1, ErrIsNull)
	}

	return nil
}

type ErrNotEqual struct {
	Expected, Actual interfaces.BinaryId
}

func IsErrNotEqual(err error) bool {
	return errors.Is(err, ErrNotEqual{})
}

func MakeErrNotEqual(expected, actual interfaces.BinaryId) (err error) {
	if Equals(expected, actual) {
		return
	}

	err = ErrNotEqual{
		Expected: expected,
		Actual:   actual,
	}

	return
}

func (err ErrNotEqual) Error() string {
	return fmt.Sprintf(
		"expected digest %s but got %s",
		err.Expected,
		err.Actual,
	)
}

func (err ErrNotEqual) Is(target error) bool {
	_, ok := target.(ErrNotEqual)
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
