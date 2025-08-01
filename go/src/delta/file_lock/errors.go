package file_lock

import (
	"fmt"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
)

type ErrLockRequired struct {
	Operation string
}

func (e ErrLockRequired) Is(target error) bool {
	_, ok := target.(ErrLockRequired)
	return ok
}

func (e ErrLockRequired) Error() string {
	return fmt.Sprintf(
		"lock required for operation: %q",
		e.Operation,
	)
}

type ErrUnableToAcquireLock struct {
	envUI       env_ui.Env
	Path        string
	description string
}

func (e ErrUnableToAcquireLock) Error() string {
	return fmt.Sprintf("%s is currently locked", e.description)
}

func (e ErrUnableToAcquireLock) Is(target error) bool {
	_, ok := target.(ErrUnableToAcquireLock)
	return ok
}

func (e ErrUnableToAcquireLock) GetHelpfulError() interfaces.ErrorHelpful {
	return e
}

func (e ErrUnableToAcquireLock) GetRetryableError() interfaces.ErrorRetryable {
	return e
}

func (e ErrUnableToAcquireLock) ErrorCause() []string {
	return []string{
		fmt.Sprintf(
			"A previous operation that acquired the %s lock failed.",
			e.description,
		),
		"The lock is intentionally left behind in case recovery is necessary.",
	}
}

func (e ErrUnableToAcquireLock) ErrorRecovery() []string {
	return []string{
		fmt.Sprintf("The lockfile needs to removed (`rm %q`).", e.Path),
	}
}

func (err ErrUnableToAcquireLock) Recover(
	ctx interfaces.RetryableContext,
	in error,
) {
	errors.PrintHelpful(err.envUI.GetErr(), err)

	if err.envUI.Confirm("delete the existing lock?") {
		if err := os.Remove(err.Path); err != nil {
			ctx.Cancel(err)
		}

		ctx.Retry()
	} else {
		errors.ContextCancelWithBadRequestf(ctx, "not deleting the lock. aborting.")
	}
}
