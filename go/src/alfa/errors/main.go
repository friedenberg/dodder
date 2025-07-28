package errors

import (
	"fmt"
	"io"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"golang.org/x/xerrors"
)

func New(text string) error {
	return xerrors.New(text)
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
		err := MakeMulti(es...)

		if err.Empty() {
			return nil
		} else {
			return err
		}
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
) (errWrapped *stackWrapError) {
	if err == io.EOF {
		panic("trying to wrap io.EOF")
	}

	if err == nil {
		return
	}

	var si stack_frame.Frame
	var ok bool

	if si, ok = stack_frame.MakeFrame(skip + 1); !ok {
		panic("failed to get stack info")
	}

	errWrapped = &stackWrapError{
		Frame: si,
	}

	if next, ok := err.(*stackWrapError); ok {
		errWrapped.next = next
	} else {
		errWrapped.error = err
	}

	if DebugBuild {
		errWrapped.checkCycle()
	}

	return
}

const thisSkip = 1

//go:noinline
func Errorf(f string, values ...any) (err error) {
	err = fmt.Errorf(f, values...)
	return
}

//go:noinline
func ErrorWithStackf(f string, values ...any) (err error) {
	err = WrapSkip(thisSkip, fmt.Errorf(f, values...))
	return
}

//go:noinline
func WrapN(n int, in error) (err error) {
	err = WrapSkip(n+thisSkip, in)
	return
}

//go:noinline
func Wrap(in error) error {
	if in == io.EOF {
		return in
	}

	return WrapSkip(thisSkip, in)
}

//go:noinline
func Wrapf(in error, format string, values ...any) error {
	if in == io.EOF {
		panic("trying to wrap io.EOF")
	}

	if in == nil {
		return nil
	}

	return &stackWrapError{
		Frame: stack_frame.MustFrame(thisSkip),
		error: fmt.Errorf(format, values...),
		next:  WrapSkip(thisSkip, in),
	}
}

//go:noinline
func IterWrapped[T any](err error) interfaces.SeqError[T] {
	return func(yield func(T, error) bool) {
		var t T
		yield(t, WrapN(1, err))
	}
}

type funcGetNext func() (error, funcGetNext)

func checkCycle(err error, next funcGetNext) {
	return
	getSlow := next
	getFast := next

	slow := err
	fast := err

	for fast != nil {
		if slow == fast && slow != err {
			panic("cycle detected!")
		}

		slow, getSlow = getSlow()
		fast, getFast = getFast()

		if fast != nil {
			fast, getFast = getFast()
		}
	}
}

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
