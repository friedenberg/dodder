package ohio_buffer

import "fmt"

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

func (e errLength) Error() string {
	return fmt.Sprintf("expected %d but got %d. error: %s", e.expected, e.actual, e.err)
}
