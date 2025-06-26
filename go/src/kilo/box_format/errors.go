package box_format

import (
	"errors"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/charlie/box"
)

type ErrBoxParse struct {
	error
}

func (err ErrBoxParse) Is(target error) bool {
	_, ok := target.(ErrBoxParse)
	return ok
}

func (err ErrBoxParse) Unwrap() error {
	return err.error
}

func (err ErrBoxParse) Error() string {
	return fmt.Sprintf("parsing box failed: %s", err.error.Error())
}

var ErrNotABox = errors.New("not a box")

type ErrBoxReadSeq struct {
	expected string
	actual   box.Seq
}

func (err ErrBoxReadSeq) Is(target error) bool {
	_, ok := target.(ErrBoxReadSeq)
	return ok
}

func (err ErrBoxReadSeq) Error() string {
	return fmt.Sprintf(
		"box parse error: expected %s but got %s",
		err.expected,
		err.actual,
	)
}
