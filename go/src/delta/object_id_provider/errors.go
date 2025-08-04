package object_id_provider

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type ErrDoesNotExist struct {
	Value string
}

func (err ErrDoesNotExist) Error() string {
	return fmt.Sprintf("object id does not exist: %s", err.Value)
}

func (err ErrDoesNotExist) Is(target error) bool {
	_, ok := target.(ErrDoesNotExist)
	return ok
}

var _ interfaces.ErrorHelpful = ErrZettelIdsExhausted{}

type ErrZettelIdsExhausted struct{}

func (err ErrZettelIdsExhausted) Error() string {
	return "zettel ids exhausted"
}

func (err ErrZettelIdsExhausted) GetErrorCause() []string {
	return []string{
		"There are no more unused zettel ids left.",
		"This may be because the last id was used.",
		"Or, it may be because this repo never had any ids to begin with.",
	}
}

func (err ErrZettelIdsExhausted) GetErrorRecovery() []string {
	return []string{
		"zettel id's must be added via the TODO command",
	}
}

func (err ErrZettelIdsExhausted) Is(target error) bool {
	_, ok := target.(ErrZettelIdsExhausted)
	return ok
}
