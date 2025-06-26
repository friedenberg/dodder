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

func (p devPrinter) Print(a ...any) (err error) {
	if !p.on {
		return
	}

	if p.includesTime {
		a = append([]any{time.Now()}, a...)
	}

	if p.includesStack {
		si, _ := stack_frame.MakeFrame(1)
		a = append([]any{si.StringNoFunctionName()}, a...)
	}

	return p.printer.Print(a...)
}

func (p devPrinter) Printf(f string, a ...any) (err error) {
	if !p.on {
		return
	}

	if p.includesTime {
		f = "%s " + f
		a = append([]any{time.Now()}, a...)
	}

	if p.includesStack {
		si, _ := stack_frame.MakeFrame(1)
		f = "%s " + f
		a = append([]any{si.StringNoFunctionName()}, a...)
	}

	return p.printer.Printf(f, a...)
}

func (p devPrinter) Caller(i int, vs ...any) {
	if !p.on {
		return
	}

	st, _ := stack_frame.MakeFrame(i + 1)

	vs = append([]any{st}, vs...)
	// TODO-P4 strip trailing newline and add back
	p.printer.Print(vs...)
}

func (p devPrinter) CallerNonEmpty(i int, v any) {
	if v != nil {
		p.Caller(i+1, "%s", v)
	}
}

func (p devPrinter) FunctionName(skip int) {
	if !p.on {
		return
	}

	st, _ := stack_frame.MakeFrame(skip + 1)
	io.WriteString(p.f, fmt.Sprintf("%s %s\n", st, st.Function))
}

//go:noinline
func (p devPrinter) Stack(skip, count int) {
	if !p.on {
		return
	}

	frames := stack_frame.MakeFrames(skip+1, count)

	io.WriteString(
		p.f,
		fmt.Sprintf(
			"Printing Stack (skip: %d, count requested: %d, count actual: %d):\n\n",
			skip,
			count,
			len(frames),
		),
	)

	for i, frame := range frames {
		io.WriteString(
			p.f,
			fmt.Sprintf("%s (%d)\n", frame.StringLogLine(), i),
		)
	}
}
