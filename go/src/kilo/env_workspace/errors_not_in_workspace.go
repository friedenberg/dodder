package env_workspace

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/hotel/workspace_config_blobs"
)

type ErrNotInWorkspace struct {
	*env
	offerToCreate bool
}

func (err ErrNotInWorkspace) Error() string {
	return "not in a workspace"
}

func (err ErrNotInWorkspace) Is(target error) bool {
	_, ok := target.(ErrNotInWorkspace)
	return ok
}

func (err ErrNotInWorkspace) ShouldShowStackTrace() bool {
	return false
}

func (err ErrNotInWorkspace) GetRetryableError() interfaces.ErrorRetryable {
	return err
}

func (err ErrNotInWorkspace) Recover(
	ctx interfaces.RetryableContext,
	in error,
) {
	if err.offerToCreate &&
		err.Confirm(
			"a workspace is necessary to run this command. create one?",
		) {
		blob := &workspace_config_blobs.V0{}

		if err := err.CreateWorkspace(blob); err != nil {
			ctx.Cancel(err)
		}

		ctx.Retry()
	} else {
		errors.ContextCancelWithBadRequestf(ctx, err.Error())
	}
}
