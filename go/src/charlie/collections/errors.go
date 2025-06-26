package collections

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var (
	ErrNilPointer        = errors.New("nil pointer")
	ErrExists            = errors.New("exists")
	MakeErrStopIteration = errors.MakeErrStopIteration
	ErrNotFound          = errNotFound("not found")
)

func MakeErrNotFound(value interfaces.Stringer) error {
	return errNotFound(value.String())
}

func MakeErrNotFoundString(s string) error {
	return errNotFound(s)
}

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound)
}

type errNotFound string

func (e errNotFound) Error() string {
	v := string(e)

	if v == "" {
		return "not found"
	} else {
		return fmt.Sprintf("not found: %q", v)
	}
}

func (e errNotFound) Is(target error) (ok bool) {
	_, ok = target.(errNotFound)
	return
}

type ErrEmptyKey[T any] struct {
	Element T
}

func (e ErrEmptyKey[T]) Error() string {
	return "empty key"
}

func (e ErrEmptyKey[T]) Is(target error) (ok bool) {
	_, ok = target.(ErrEmptyKey[T])
	return
}
