package ui

import "fmt"

func MakePrefixPrinter(printer Printer, prefix string) prefixPrinter {
	return prefixPrinter{
		Printer: printer,
		prefix:  prefix,
	}
}

type prefixPrinter struct {
	Printer
	prefix string
}

func (printer prefixPrinter) Print(v ...any) error {
	return printer.Printer.Print(
		append([]any{printer.prefix}, v...)...,
	)
}

func (printer prefixPrinter) Printf(format string, v ...any) error {
	return printer.Printer.Printf(
		fmt.Sprintf("%s%s", printer.prefix, format),
		v...,
	)
}
