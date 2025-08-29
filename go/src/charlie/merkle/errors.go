package merkle

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var ErrEmptyType = errors.New("type is empty")

func MakeErrEmptyType(id interfaces.BlobId) error {
	if id.GetType() == "" {
		return errors.WrapSkip(1, ErrEmptyType)
	}

	return nil
}

func MakeErrIsNotNull(id interfaces.BlobId) error {
	if !id.IsNull() {
		// TODO clone
		return errors.WrapSkip(1, errIsNotNull{id: id, value: id.String()})
	}

	return nil
}

type errIsNotNull struct {
	value string
	id    interfaces.BlobId
}

func (err errIsNotNull) Error() string {
	return fmt.Sprintf("blob id is not null %q", err.value)
}

func (err errIsNotNull) Is(target error) bool {
	_, ok := target.(errIsNotNull)
	return ok
}

func MakeErrIsNull(id interfaces.BlobId, key string) error {
	if id.IsNull() {
		return errors.WrapSkip(1, errIsNull{key: key})
	}

	return nil
}

type errIsNull struct {
	key string
}

func (err errIsNull) Error() string {
	return fmt.Sprintf("blob id is null for key %q", err.key)
}

func (err errIsNull) Is(target error) bool {
	_, ok := target.(errIsNull)
	return ok
}

type ErrNotEqual struct {
	Expected, Actual interfaces.BlobId
}

func MakeErrNotEqual(expected, actual interfaces.BlobId) (err error) {
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

type ErrNotEqualBytes struct {
	Expected, Actual []byte
}

func MakeErrNotEqualBytes(expected, actual []byte) (err error) {
	if bytes.Equal(expected, actual) {
		return
	}

	err = ErrNotEqualBytes{
		Expected: expected,
		Actual:   actual,
	}

	return
}

func (err ErrNotEqualBytes) Error() string {
	return fmt.Sprintf(
		"expected digest %x but got %x",
		err.Expected,
		err.Actual,
	)
}

func (err ErrNotEqualBytes) Is(target error) bool {
	_, ok := target.(ErrNotEqualBytes)
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

func (err errLength) Error() string {
	return fmt.Sprintf(
		"expected digest to have length %d, but got %d",
		err[0],
		err[1],
	)
}
