package errors

import (
	"errors"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
)

type (
	Typed[DISAMB any] interface {
		error
		GetErrorType() DISAMB
	}

	errorString[DISAMB any] struct {
		value string
	}

	errorTypedWrapped[DISAMB any] struct {
		wrapped error
	}
)

func IsTyped[DISAMB any](err error) bool {
	var target *errorString[DISAMB]
	return Is(err, target)
}

func NewWithType[DISAMB any](text string) Typed[DISAMB] {
	return &errorString[DISAMB]{text}
}

func WrapWithType[DISAMB any](err error) Typed[DISAMB] {
	return &errorTypedWrapped[DISAMB]{wrapped: err}
}

func (err *errorTypedWrapped[TYPE]) Error() string {
	return err.wrapped.Error()
}

func (err *errorTypedWrapped[TYPE]) GetErrorType() TYPE {
	var disamb TYPE
	return disamb
}

func (err *errorTypedWrapped[_]) Unwrap() error {
	return err.wrapped
}

func (err *errorString[_]) Error() string {
	return err.value
}

func (err *errorString[TYPE]) GetErrorType() TYPE {
	var disamb TYPE
	return disamb
}

func (err *errorString[DISAMB]) Is(target error) bool {
	_, ok := target.(*errorString[DISAMB])
	return ok
}

// Stop iteration sentinel
type errStopIterationDisamb struct{}

var errStopIteration = NewWithType[errStopIterationDisamb]("stop iteration")

func MakeErrStopIteration() error {
	return errStopIteration
}

func IsStopIteration(err error) bool {
	ok := Is(err, errStopIteration)
	return ok
}

// Exists sentinel
type errExistsDisamb struct{}

var ErrExists = NewWithType[errExistsDisamb]("exists")

// Not found sentinel
type errNotFoundDisamb struct{}

type ErrNotFound struct {
	Value string
}

func MakeErrNotFound(value interfaces.Stringer) error {
	return ErrNotFound{Value: value.String()}
}

func MakeErrNotFoundString(s string) error {
	return ErrNotFound{Value: s}
}

func IsErrNotFound(err error) bool {
	return errors.Is(err, ErrNotFound{})
}

func (err ErrNotFound) Error() string {
	if err.Value == "" {
		return "not found"
	} else {
		return fmt.Sprintf("not found: %q", err.Value)
	}
}

func (err ErrNotFound) Is(target error) (ok bool) {
	_, ok = target.(ErrNotFound)
	return ok
}

func (err ErrNotFound) GetErrorType() errNotFoundDisamb {
	return errNotFoundDisamb{}
}
