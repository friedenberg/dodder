package errors

import (
	"errors"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

// TODO redesign all the below
var errStopIteration = New("stop iteration")

func MakeErrStopIteration() error {
	return errStopIteration
}

func IsStopIteration(err error) bool {
	ok := Is(err, errStopIteration)
	return ok
}

type (
	errorString[DISAMB any] struct {
		value string
	}
)

func IsTyped[DISAMB any](err error) bool {
	var target *errorString[DISAMB]
	return Is(err, target)
}

func New(text string) error {
	return &errorString[string]{text}
}

func NewWithType[DISAMB any](text string) error {
	return &errorString[DISAMB]{text}
}

func (err *errorString[_]) Error() string {
	return err.value
}

func (err *errorString[DISAMB]) Is(target error) bool {
	_, ok := target.(*errorString[DISAMB])
	return ok
}

var ErrExists = New("exists")

func MakeErrNotFound(value interfaces.Stringer) error {
	return ErrNotFound(value.String())
}

func MakeErrNotFoundString(s string) error {
	return ErrNotFound(s)
}

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound(""))
}

type ErrNotFound string

func (err ErrNotFound) Error() string {
	v := string(err)

	if v == "" {
		return "not found"
	} else {
		return fmt.Sprintf("not found: %q", v)
	}
}

func (err ErrNotFound) Is(target error) (ok bool) {
	_, ok = target.(ErrNotFound)
	return ok
}
