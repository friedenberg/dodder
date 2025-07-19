package interfaces

import (
	"context"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

//go:generate stringer -type=ContextState
type ContextState uint8

const (
	ContextStateUnknown = ContextState(iota)
	ContextStateUnstarted
	// all states that are > than `ContextStateStarted` are considered terminal,
	// so the order here is important
	ContextStateStarted
	ContextStateSucceeded
	ContextStateFailed
	ContextStateAborted
)

func (state ContextState) IsComplete() bool {
	return state > ContextStateStarted
}

func (state ContextState) Error() string {
	return state.String()
}

func (state ContextState) Is(target error) bool {
	_, ok := target.(ContextState)
	return ok
}

type (
	FuncContext = func(Context) error

	// TODO think about how to separate "consumers" of context, and "managers or
	// supervisors"

	ActiveContext interface {
		context.Context

		Cause() error
		GetState() ContextState

		// TODO disambiguate between errors and exceptions
		// TODO rename this to Complete
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
		CauseWithStackFrames() (error, []stack_frame.Frame)
		Run(func(Context)) error

		// TODO extricate from *context and turn into generic function
		SetCancelOnSignals(signals ...os.Signal)
	}

	RetryableContext interface {
		Context
		Retry()
	}
)
