package env_workspace

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/papa/store_workspace"
)

type ErrUnsupportedType struct {
	Type ids.Type
}

func (err ErrUnsupportedType) Is(target error) bool {
	_, ok := target.(ErrUnsupportedType)
	return ok
}

func (err ErrUnsupportedType) Error() string {
	return fmt.Sprintf("unsupported type: %q", err.Type)
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
