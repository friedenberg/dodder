package errors

import "code.linenisgreat.com/dodder/go/src/alfa/stack_frame"

type Flusher interface {
	Flush() error
}

type (
	Func func() error
)

type FuncWithStackInfo struct {
	Func
	stack_frame.Frame
}

type WithStackInfo[T any] struct {
	Contents T
	stack_frame.Frame
}

type WaitGroup interface {
	Do(Func) bool
	DoAfter(Func)
	GetError() error
}

func MakeNilFunc(in func()) Func {
	return func() error {
		in()
		return nil
	}
}
