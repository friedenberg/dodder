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

	// When printing error trees, `error_coders` uses the presence of
	// `ShouldHideUnwrap()` and its return value to determine if the parent
	// error
	// should be printed.
	ErrorHiddenWrapper interface {
		ErrorOneUnwrapper
		ShouldHideUnwrap() bool
	}

	ErrorHelpful interface {
		error
		GetErrorCause() []string
		GetErrorRecovery() []string
	}
)
