package ui

import "fmt"

func MakePrefixPrinter(printer printer, prefix string) prefixPrinter {
	return prefixPrinter{
		printer: printer,
		prefix:  prefix,
	}
}

type prefixPrinter struct {
	printer
	prefix string
}

func (printer prefixPrinter) GetPrinter() Printer {
	return printer
}

func (printer prefixPrinter) Print(v ...interface{}) error {
	return printer.printer.Print(
		append([]interface{}{printer.prefix}, v...)...,
	)
}

func (printer prefixPrinter) Printf(format string, v ...interface{}) error {
	return printer.printer.Printf(
		fmt.Sprintf("%s%s", printer.prefix, format),
		v...,
	)
}
