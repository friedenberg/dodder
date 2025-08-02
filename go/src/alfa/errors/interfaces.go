package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type (
	Flusher interface {
		Flush() error
	}

	FuncNil     = func()
	FuncErr     = func() error
	FuncContext = interfaces.FuncContext

	FuncWithStackInfo struct {
		FuncErr
		stack_frame.Frame
	}

	WithStackInfo[T any] struct {
		Contents T
		stack_frame.Frame
	}

	WaitGroup interface {
		Do(FuncErr) bool
		DoAfter(FuncErr)
		GetError() error
	}

	FuncIs func(error) bool

	UnwrapOne  = interfaces.ErrorOneUnwrapper
	UnwrapMany = interfaces.ErrorManyUnwrapper

	ErrorsIs interface {
		Is(error) bool
	}
)

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
