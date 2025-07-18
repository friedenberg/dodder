package interfaces

import (
	"context"
	"os"
)

type (
	FuncContext = func(Context) error

	ActiveContext interface {
		context.Context

		Cause() error
		// returns true if the context is still active, or false if it's been
		// cancelled for any reason
		Continue() bool
		// TODO disambiguate between errors and exceptions
		Cancel(error)

		// `After` runs a function after the context is complete (regardless of
		// any errors). `After`s are run in the reverse order of when they are
		// called,
		// like
		// defers but on a whole-program level.
		After(FuncContext)

		// `Must` executes a function even if the context has been cancelled. If
		// the function returns an error, `Must` cancels the context. It is
		// meant for
		// defers that must be executed, like closing files, flushing buffers,
		// releasing locks.
		Must(FuncContext)
	}

	Context interface {
		ActiveContext
		Run(func(Context)) error
		SetCancelOnSignals(signals ...os.Signal)
	}

	RetryableContext interface {
		Context
		Retry()
	}
)
