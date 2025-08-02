package errors

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

func WithHelp(
	err error,
	cause []string,
	recovery []string,
) interfaces.ErrorHelpful {
	return helpful{
		underlying: err,
		cause:      cause,
		recovery:   recovery,
	}
}

type helpful struct {
	underlying error
	cause      []string
	recovery   []string
}

func (err helpful) Error() string {
	return err.underlying.Error()
}

func (err helpful) Unwrap() error {
	return err.underlying
}

func (err helpful) ErrorCause() []string {
	return err.cause
}

func (err helpful) ErrorRecovery() []string {
	return err.recovery
}

func (err helpful) ShouldShowStackTrace() bool {
	return false
}

func (err helpful) ShouldHideUnwrap() bool {
	return true
}
