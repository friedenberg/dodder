package doddish

import (
	"errors"
	"fmt"
)

var ErrEmptySeq = errors.New("empty seq")

type ErrUnsupportedSeq struct {
	Seq
}

func (err ErrUnsupportedSeq) Error() string {
	return fmt.Sprintf("unsupported seq: %q", err.Seq)
}

func (err ErrUnsupportedSeq) Is(target error) bool {
	_, ok := target.(ErrUnsupportedSeq)
	return ok
}
