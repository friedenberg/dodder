package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type (
	errTypeLockDisamb struct{}
)

func newTypeLockError(text string) error {
	return errors.NewWithType[errTypeLockDisamb](text)
}

func IsTypeLockError(err error) bool {
	return errors.IsTyped[errTypeLockDisamb](err)
}

var (
	ErrFailedToReadCurrentLockObject = newTypeLockError("failed to read current lock object")
	ErrEmptyLockKey                  = newTypeLockError("empty type")
	ErrBuiltinType                   = newTypeLockError("builtin type")
)
