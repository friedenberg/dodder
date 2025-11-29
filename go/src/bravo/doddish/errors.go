package doddish

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type (
	pkgErrDisamb struct{}
)

func newPkgError(text string) error {
	return errors.NewWithType[pkgErrDisamb](text)
}

var ErrEmptySeq = newPkgError("empty seq")

type ErrUnsupportedSeq struct {
	Seq
}

func (err ErrUnsupportedSeq) Error() string {
	return fmt.Sprintf("unsupported seq: %q", err.Seq)
}

func (err ErrUnsupportedSeq) Is(target error) bool {
	if _, ok := target.(ErrUnsupportedSeq); ok {
		return true
	}

	return false
}
