package toml

import (
	"github.com/pelletier/go-toml/v2"
)

type (
	StrictMissingError = toml.StrictMissingError
)

type Error struct {
	error
}

func MakeError(err error) Error {
	return Error{
		error: err,
	}
}

func (err Error) Is(target error) (ok bool) {
	_, ok = target.(Error)
	return
}
