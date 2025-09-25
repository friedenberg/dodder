package ui

import (
	"fmt"
	"os"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/primordial"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

func MakePrinter(file *os.File) printer {
	return MakePrinterOn(file, true)
}

func MakePrinterOn(file *os.File, on bool) printer {
	return printer{
		file:  file,
		isTty: primordial.IsTty(file),
		on:    on,
	}
}

type printer struct {
	file  *os.File
	isTty bool
	on    bool
}

// Returns a copy of this printer with a modified `on` setting
func (printer printer) withOn(on bool) printer {
	printer.on = on
	return printer
}

func (printer printer) GetPrinter() Printer {
	return printer
}

func (printer printer) Write(b []byte) (n int, err error) {
	if !printer.on {
		n = len(b)
		return n, err
	}

	return printer.file.Write(b)
}

func (printer printer) GetFile() *os.File {
	return printer.file
}

func (printer printer) IsTty() bool {
	return printer.isTty
}

//go:noinline
func (printer printer) Caller(skip int) Printer {
	if !printer.on {
		return Null
	}

	stackFrame, _ := stack_frame.MakeFrame(skip + 1)

	return prefixPrinter{
		Printer: printer,
		prefix:  stackFrame.StringNoFunctionName() + " ",
	}
}

func (printer printer) PrintDebug(args ...any) (err error) {
	if !printer.on {
		return err
	}

	_, err = fmt.Fprintf(
		printer.file,
		strings.Repeat("%#v ", len(args))+"\n",
		args...,
	)

	return err
}

func (printer printer) Print(args ...any) (err error) {
	if !printer.on {
		return err
	}

	_, err = fmt.Fprintln(
		printer.file,
		args...,
	)

	return err
}

//go:noinline
func (printer printer) printfStack(
	depth int,
	format string,
	args ...any,
) (err error) {
	if !printer.on {
		return err
	}

	stackFrame, _ := stack_frame.MakeFrame(1 + depth)
	format = "%s" + format
	args = append([]any{stackFrame}, args...)

	_, err = fmt.Fprintln(
		printer.file,
		fmt.Sprintf(format, args...),
	)

	return err
}

func (printer printer) Printf(format string, args ...any) (err error) {
	if !printer.on {
		return err
	}

	_, err = fmt.Fprintln(
		printer.file,
		fmt.Sprintf(format, args...),
	)

	return err
}
