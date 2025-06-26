package catgut

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

var ErrBufferEmpty = errors.New("buffer empty")

type errLength struct {
	expected, actual int64
	err              error
}

func MakeErrLength(expected, actual int64, err error) error {
	switch {
	case expected != actual:
		return errLength{expected, actual, err}

	case err != nil:
		return errLength{expected, actual, err}

	default:
		return nil
	}
}

func (a errLength) Is(b error) (ok bool) {
	_, ok = b.(errLength)
	return
}

func (e errLength) Error() string {
	return fmt.Sprintf("expected %d but got %d. error: %s", e.expected, e.actual, e.err)
}

type errInvalidSliceRange [2]int

func (a errInvalidSliceRange) Is(b error) (ok bool) {
	_, ok = b.(errInvalidSliceRange)
	return
}

func (e errInvalidSliceRange) Error() string {
	return fmt.Sprintf("invalid slice range: (%d, %d)", e[0], e[1])
}
