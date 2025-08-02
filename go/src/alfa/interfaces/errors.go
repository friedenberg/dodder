package interfaces

import "code.linenisgreat.com/dodder/go/src/alfa/stack_frame"

type (
	ErrorStackTracer = stack_frame.ErrorStackTracer

	ErrorOneUnwrapper interface {
		error
		Unwrap() error
	}

	ErrorManyUnwrapper interface {
		error
		Unwrap() []error
	}

	ErrorHiddenWrapper interface {
		ErrorOneUnwrapper
		ShouldHideUnwrap() bool
	}

	ErrorBadRequest interface {
		IsBadRequest()
	}

	ErrorHelpful interface {
		error
		// TODO prefix with Get
		ErrorCause() []string
		ErrorRecovery() []string
	}

	ErrorRetryable interface {
		error
		Recover(RetryableContext, error)
	}
)
