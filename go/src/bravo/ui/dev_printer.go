package ui

import (
	"fmt"
	"io"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

type devPrinter struct {
	printer
	includesTime  bool
	includesStack bool
}

//go:noinline
func (printer devPrinter) Print(args ...any) (err error) {
	if !printer.on {
		return
	}

	if printer.includesTime {
		args = append([]any{time.Now()}, args...)
	}

	if printer.includesStack {
		stackFrame, _ := stack_frame.MakeFrame(1)
		args = append([]any{stackFrame.StringNoFunctionName()}, args...)
	}

	return printer.printer.Print(args...)
}

//go:noinline
func (printer devPrinter) Printf(format string, args ...any) (err error) {
	if !printer.on {
		return
	}

	if printer.includesTime {
		format = "%s " + format
		args = append([]any{time.Now()}, args...)
	}

	if printer.includesStack {
		stackFrame, _ := stack_frame.MakeFrame(1)
		format = "%s " + format
		args = append([]any{stackFrame.StringNoFunctionName()}, args...)
	}

	return printer.printer.Printf(format, args...)
}

//go:noinline
func (printer devPrinter) Caller(skip int, args ...any) {
	if !printer.on {
		return
	}

	stackFrame, _ := stack_frame.MakeFrame(skip + 1)

	args = append([]any{stackFrame}, args...)
	// TODO-P4 strip trailing newline and add back
	printer.printer.Print(args...)
}

//go:noinline
func (printer devPrinter) CallerNonEmpty(skip int, arg any) {
	if arg != nil {
		printer.Caller(skip+1, "%s", arg)
	}
}

//go:noinline
func (printer devPrinter) FunctionName(skip int) {
	if !printer.on {
		return
	}

	stackFrame, _ := stack_frame.MakeFrame(skip + 1)
	io.WriteString(
		printer.file,
		fmt.Sprintf("%s %s\n", stackFrame, stackFrame.Function),
	)
}

//go:noinline
func (printer devPrinter) Stack(skip, count int) {
	if !printer.on {
		return
	}

	frames := stack_frame.MakeFrames(skip+1, count)

	io.WriteString(
		printer.file,
		fmt.Sprintf(
			"Printing Stack (skip: %d, count requested: %d, count actual: %d):\n\n",
			skip,
			count,
			len(frames),
		),
	)

	for i, frame := range frames {
		io.WriteString(
			printer.file,
			fmt.Sprintf("%s (%d)\n", frame.StringLogLine(), i),
		)
	}
}
