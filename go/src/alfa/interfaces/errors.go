package interfaces

type (
	ErrorStackTracer interface {
		error
		ShouldShowStackTrace() bool
	}

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
