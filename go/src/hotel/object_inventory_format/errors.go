package object_inventory_format

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type errInvalidGenericFormat string

func (err errInvalidGenericFormat) Error() string {
	return fmt.Sprintf("invalid format: %q", string(err))
}

func (err errInvalidGenericFormat) Is(target error) bool {
	_, ok := target.(errInvalidGenericFormat)
	return ok
}

var ErrEmptyTai = errors.New("empty tai")
