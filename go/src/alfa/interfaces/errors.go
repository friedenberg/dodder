package interfaces

import "code.linenisgreat.com/dodder/go/src/alfa/stack_frame"

type (
	ErrorStackTracer = stack_frame.ErrorStackTracer

	ErrorBadRequest interface {
		IsBadRequest()
	}

	ErrorHelpful interface {
		error
		GetHelpfulError() ErrorHelpful
		ErrorCause() []string
		ErrorRecovery() []string
	}

	ErrorRetryable interface {
		GetRetryableError() ErrorRetryable
		Recover(RetryableContext, error)
	}
)
