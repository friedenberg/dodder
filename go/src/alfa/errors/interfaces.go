package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type Flusher interface {
	Flush() error
}

type (
	FuncNil     = func()
	FuncErr     = func() error
	FuncContext = interfaces.FuncContext
)

type FuncWithStackInfo struct {
	FuncErr
	stack_frame.Frame
}

type WithStackInfo[T any] struct {
	Contents T
	stack_frame.Frame
}

type WaitGroup interface {
	Do(FuncErr) bool
	DoAfter(FuncErr)
	GetError() error
}

func MakeFuncErrFromFuncNil(in FuncNil) FuncErr {
	return func() error {
		in()
		return nil
	}
}

func MakeFuncContextFromFuncErr(in FuncErr) FuncContext {
	return func(interfaces.Context) error {
		return in()
	}
}

func MakeFuncContextFromFuncNil(in FuncNil) FuncContext {
	return func(interfaces.Context) error {
		in()
		return nil
	}
}
