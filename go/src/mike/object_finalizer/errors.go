package object_finalizer

import "errors"

var (
	ErrFailedToReadCurrentTypeObject = errors.New("failed to read current type object")
	ErrEmptyType                     = errors.New("empty type")
	ErrBuiltinType                   = errors.New("builtin type")
)
