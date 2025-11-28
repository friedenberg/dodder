package object_finalizer

import "code.linenisgreat.com/dodder/go/src/alfa/errors"

type (
	errTypeLockfileDisamb struct{}
	errTypeLockfile       = errors.TypedError[errTypeLockfileDisamb]
)

func newTypeLockfileError(text string) errTypeLockfile {
	return errors.NewWithType[errTypeLockfileDisamb](text)
}

var (
	ErrFailedToReadCurrentTypeObject = newTypeLockfileError("failed to read current type object")
	ErrEmptyType                     = newTypeLockfileError("empty type")
	ErrBuiltinType                   = newTypeLockfileError("builtin type")
)
