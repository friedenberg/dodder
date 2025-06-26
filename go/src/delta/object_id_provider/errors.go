package object_id_provider

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type ErrDoesNotExist struct {
	Value string
}

func (e ErrDoesNotExist) Error() string {
	return fmt.Sprintf("object id does not exist: %s", e.Value)
}

func (e ErrDoesNotExist) Is(target error) bool {
	_, ok := target.(ErrDoesNotExist)
	return ok
}

type ErrZettelIdsExhausted struct{}

func (e ErrZettelIdsExhausted) GetHelpfulError() errors.Helpful {
	return e
}

func (e ErrZettelIdsExhausted) Error() string {
	return "zettel ids exhausted"
}

func (e ErrZettelIdsExhausted) ErrorCause() []string {
	return []string{
		"There are no more unused zettel ids left.",
		"This may be because the last id was used.",
		"Or, it may be because this repo never had any ids to begin with.",
	}
}

func (e ErrZettelIdsExhausted) ErrorRecovery() []string {
	return []string{
		"zettel id's must be added via the TODO command",
	}
}

func (e ErrZettelIdsExhausted) Is(target error) bool {
	_, ok := target.(ErrZettelIdsExhausted)
	return ok
}
