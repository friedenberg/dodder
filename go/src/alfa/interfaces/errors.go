package interfaces

type (
	StackTracer interface {
		error
		ShouldShowStackTrace() bool
	}

	ErrBadRequest interface {
		IsBadRequest()
	}
)
