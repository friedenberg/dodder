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
	For string
	Seq
}

func (err ErrUnsupportedSeq) Error() string {
	if err.For == "" {
		return fmt.Sprintf("unsupported seq: %q", err.Seq)
	} else {
		return fmt.Sprintf("unsupported seq %q for %q", err.Seq, err.For)
	}
}

func (err ErrUnsupportedSeq) Is(target error) bool {
	if _, ok := target.(ErrUnsupportedSeq); ok {
		return true
	}

	return false
}
