package env_workspace

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

type ErrUnsupportedType ids.Type

func (e ErrUnsupportedType) Is(target error) bool {
	_, ok := target.(ErrUnsupportedType)
	return ok
}

func (e ErrUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported typ: %q", ids.Type(e))
}

func makeErrUnsupportedOperation(s *Store, op any) error {
	return ErrUnsupportedOperation{
		repoId:             s.RepoId,
		store:              s.StoreLike,
		operationInterface: op,
	}
}

type ErrUnsupportedOperation struct {
	repoId             ids.RepoId
	store              store_workspace.StoreLike
	operationInterface any
}

func (e ErrUnsupportedOperation) Error() string {
	return fmt.Sprintf(
		"store (%q:%T) does not support operation '%T'",
		e.repoId,
		e.store,
		e.operationInterface,
	)
}

func (e ErrUnsupportedOperation) Is(target error) bool {
	_, ok := target.(ErrUnsupportedOperation)
	return ok
}
