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

func (err ErrLockRequired) Is(target error) bool {
	_, ok := target.(ErrLockRequired)
	return ok
}

func (err ErrLockRequired) Error() string {
	return fmt.Sprintf(
		"lock required for operation: %q",
		err.Operation,
	)
}

type ErrUnableToAcquireLock struct {
	envUI       env_ui.Env
	Path        string
	description string
}

var _ interfaces.ErrorRetryable = ErrUnableToAcquireLock{}

func (err ErrUnableToAcquireLock) Error() string {
	return fmt.Sprintf("%s is currently locked", err.description)
}

func (err ErrUnableToAcquireLock) Is(target error) bool {
	_, ok := target.(ErrUnableToAcquireLock)
	return ok
}

func (err ErrUnableToAcquireLock) GetErrorCause() []string {
	return []string{
		fmt.Sprintf(
			"A previous operation that acquired the %s lock failed.",
			err.description,
		),
		"The lock is intentionally left behind in case recovery is necessary.",
	}
}

func (err ErrUnableToAcquireLock) GetErrorRecovery() []string {
	return []string{
		fmt.Sprintf("The lockfile needs to removed (`rm %q`).", err.Path),
	}
}

func (err ErrUnableToAcquireLock) Recover(
	ctx interfaces.ActiveContext,
	retry interfaces.FuncRetry,
	abort interfaces.FuncRetryAborted,
) {
	errors.PrintHelpful(err.envUI.GetErr(), err)

	if err.envUI.Confirm("delete the existing lock?", "") {
		if err := os.Remove(err.Path); err != nil {
			ctx.Cancel(err)
		}

		retry()
	} else {
		abort(errors.Errorf("not deleting the lock"))
	}
}
