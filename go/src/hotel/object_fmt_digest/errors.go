package object_fmt_digest

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

var ErrEmptyTai = errors.New("empty tai")

type errUnknownFormatKey string

func (err errUnknownFormatKey) Error() string {
	return fmt.Sprintf("unknown format key: %q", string(err))
}

func (err errUnknownFormatKey) Is(target error) bool {
	_, ok := target.(errUnknownFormatKey)
	return ok
}
