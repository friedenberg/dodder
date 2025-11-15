package collections

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

var ErrExists = errors.New("exists")

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
