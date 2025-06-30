package ui

import (
	"fmt"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

// TODO move to todo_debug and replace with no-op stubs
var todo todoPrinter

func init() {
	todo = todoPrinter{
		printer: printer{
			f: os.Stderr,
		},
		includesStack: true,
	}
}

func SetTodoOn() {
	todo.on = true
}

type todoPrinter devPrinter

//go:generate stringer -type=Priority
type Priority int

const (
	P0 = Priority(iota)
	P1
	P2
	P3
	P4
	P5
)

func TodoRecoverable(f string, a ...any) (err error) {
	return printerErr.printfStack(1, "TODO: Make recoverable: "+f, a...)
}

func Todo(f string, a ...any) (err error) {
	return printerErr.printfStack(1, "TODO: "+f, a...)
}

func TodoP0(f string, a ...any) (err error) {
	return todo.printf(1, P0, f, a...)
}

func TodoP1(f string, a ...any) (err error) {
	return todo.printf(1, P1, f, a...)
}

func TodoP2(f string, a ...any) (err error) {
	return todo.printf(1, P2, f, a...)
}

func TodoP3(f string, a ...any) (err error) {
	return todo.printf(1, P3, f, a...)
}

func TodoP4(f string, a ...any) (err error) {
	return todo.printf(1, P4, f, a...)
}

func TodoP5(f string, a ...any) (err error) {
	return todo.printf(1, P5, f, a...)
}

func (p todoPrinter) Printf(
	pr Priority,
	f string,
	a ...any,
) (err error) {
	return p.printf(1, pr, f, a...)
}

func (p todoPrinter) printf(
	skip int,
	pr Priority,
	f string,
	a ...any,
) (err error) {
	if !p.on {
		return
	}

	if p.includesStack {
		si, _ := stack_frame.MakeFrame(1 + skip)
		f = "%s %s" + f
		a = append([]any{pr, si}, a...)
	}

	_, err = fmt.Fprintln(
		p.f,
		fmt.Sprintf(f, a...),
	)

	return
}
