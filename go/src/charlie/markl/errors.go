package markl

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"golang.org/x/exp/constraints"
)

var ErrEmptyType = errors.New("type is empty")

func MakeErrEmptyType(id interfaces.MarklId) error {
	if id.GetMarklFormat() == nil {
		return errors.WrapSkip(1, ErrEmptyType)
	}

	return nil
}

func AssertIdIsNull(id interfaces.MarklId) error {
	if !id.IsNull() {
		// TODO clone
		return errors.WrapSkip(1, errIsNotNull{id: id, value: id.String()})
	}

	return nil
}

type errIsNotNull struct {
	value string
	id    interfaces.MarklId
}

func (err errIsNotNull) Error() string {
	return fmt.Sprintf("blob id is not null %q", err.value)
}

func (err errIsNotNull) Is(target error) bool {
	_, ok := target.(errIsNotNull)
	return ok
}

// TODO remove key
func AssertIdIsNotNull(id interfaces.MarklId, key string) error {
	format := id.GetPurpose()

	if format != "" {
		key = format
	}

	if id.IsNull() {
		return errors.WrapSkip(1, errIsNull{key: key})
	}

	return nil
}

func IsErrNull(target error) bool {
	return errors.Is(target, errIsNull{})
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
	Expected, Actual interfaces.MarklId
}

func AssertEqual(expected, actual interfaces.MarklId) (err error) {
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
		"expected digest %q but got %q",
		err.Expected,
		err.Actual,
	)
}

func (err ErrNotEqual) Is(target error) bool {
	_, ok := target.(ErrNotEqual)
	return ok
}

func (err ErrNotEqual) IsDifferentHashTypes() bool {
	return err.Expected.GetMarklFormat() != err.Actual.GetMarklFormat()
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

type errLength[INTEGER constraints.Integer] [2]INTEGER

// TODO add another "wrong hasher" error type
func MakeErrLength[INTEGER constraints.Integer](
	expected, actual INTEGER,
) error {
	if expected != actual {
		return errLength[INTEGER]{expected, actual}
	} else {
		return nil
	}
}

func (err errLength[_]) Error() string {
	return fmt.Sprintf(
		"expected digest to have length %d, but got %d",
		err[0],
		err[1],
	)
}

func MakeErrWrongType(expected, actual string) error {
	if expected != actual {
		return errWrongType{expected: expected, actual: actual}
	}

	return nil
}

type errWrongType struct {
	expected, actual string
}

func (err errWrongType) Error() string {
	return fmt.Sprintf(
		"wrong type. expected %q but got %q",
		err.expected,
		err.actual,
	)
}

func (err errWrongType) Is(target error) bool {
	_, ok := target.(errWrongType)
	return ok
}
