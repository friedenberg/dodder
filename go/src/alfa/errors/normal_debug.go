//go:build debug

package errors

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

func PrintStackTracerIfNecessary(
	printer interfaces.Printer,
	name string,
	err error,
) {
	printer.Printf("\n\n%s failed with error:\n%s", name, err)
}
