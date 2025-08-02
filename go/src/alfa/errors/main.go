package errors

import (
	"errors"
	"fmt"
	"io"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

func New(text string) error {
	return errors.New(text)
}

func Join(es ...error) error {
	switch {
	case len(es) == 2 && es[0] == nil && es[1] == nil:
		return nil

	case len(es) == 2 && es[0] == nil:
		return es[1]

	case len(es) == 2 && es[1] == nil:
		return es[0]

	default:
		groupBuilder := MakeGroupBuilder(es...)
		return groupBuilder.GetError()
	}
}

func PanicIfError(err any) {
	if err == nil {
		return
	}

	switch t := err.(type) {
	case func() error:
		PanicIfError(t())

	case error:
		panic(t)
	}
}

//go:noinline
func WrapSkip(
	skip int,
	err error,
) error {
	if shouldNotWrap(err) {
		return err
	}

	var stackFrame stack_frame.Frame
	var ok bool

	if stackFrame, ok = stack_frame.MakeFrame(skip + 1); !ok {
		panic("failed to get stack info")
	}

	return stackFrame.Wrap(err)
}

const thisSkip = 1

//go:noinline
func Errorf(f string, values ...any) (err error) {
	err = fmt.Errorf(f, values...)
	return
}

//go:noinline
func ErrorWithStackf(format string, args ...any) error {
	var stackFrame stack_frame.Frame
	var ok bool

	if stackFrame, ok = stack_frame.MakeFrame(thisSkip); !ok {
		panic("failed to get stack info")
	}

	return stackFrame.Errorf(format, args...)
}

func shouldNotWrap(err error) bool {
	switch err {
	case nil:
		return true

	case io.EOF:
		panic("trying to wrap io.EOF")
	}

	var errStackTracer interfaces.ErrorStackTracer

	if As(err, &errStackTracer) && !errStackTracer.ShouldShowStackTrace() {
		return true
	} else {
		return false
	}
}

//go:noinline
func Wrap(err error) error {
	if shouldNotWrap(err) {
		return err
	}

	var stackFrame stack_frame.Frame
	var ok bool

	if stackFrame, ok = stack_frame.MakeFrame(thisSkip); !ok {
		panic("failed to get stack info")
	}

	return stackFrame.Wrap(err)
}

//go:noinline
func Wrapf(err error, format string, values ...any) error {
	if shouldNotWrap(err) {
		return err
	}

	var stackFrame stack_frame.Frame
	var ok bool

	if stackFrame, ok = stack_frame.MakeFrame(thisSkip); !ok {
		panic("failed to get stack info")
	}

	return stackFrame.Wrapf(err, format, values...)
}

//go:noinline
func IterWrapped[T any](err error) interfaces.SeqError[T] {
	return func(yield func(T, error) bool) {
		var t T
		yield(t, WrapSkip(1, err))
	}
}

type funcGetNext func() (error, funcGetNext)

// TODO remove / rewrite the below

// Wrap the error with stack info unless it's one of the provided `except`
// errors, in which case return nil. Direct value comparison is
// performed (`in == except`) rather than errors.Is.
//
//go:noinline
func WrapExceptSentinelAsNil(in error, except ...error) (err error) {
	if in == nil {
		return
	}

	if slices.Contains(except, in) {
		return in
	}

	err = WrapSkip(thisSkip, in)

	return
}

// Wrap the error with stack info unless it's one of the provided `except`
// errors, in which case return that bare error. Direct value comparison is
// performed (`in == except`) rather than errors.Is.
//
//go:noinline
func WrapExceptSentinel(in error, except ...error) (err error) {
	if in == nil {
		return
	}

	if slices.Contains(except, in) {
		return in
	}

	err = WrapSkip(thisSkip, in)

	return
}
