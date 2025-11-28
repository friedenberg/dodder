package doddish

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type (
	errDisamb struct{}
	pkgError  = errors.TypedError[errDisamb]
)

func newError(text string) pkgError {
	return errors.NewWithType[errDisamb](text)
}

var ErrEmptySeq = newError("empty seq")

type ErrUnsupportedSeq struct {
	pkgError
	Seq
}

func (err ErrUnsupportedSeq) Error() string {
	return fmt.Sprintf("unsupported seq: %q", err.Seq)
}

func (err ErrUnsupportedSeq) Is(target error) bool {
	_, ok := target.(ErrUnsupportedSeq)
	return ok
}
