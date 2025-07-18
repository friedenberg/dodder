//go:build !debug

package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func PrintStackTracerIfNecessary(
	printer interfaces.Printer,
	name string,
	err error,
	_ ...interface{},
) {
	var normalError StackTracer

	if As(err, &normalError) && !normalError.ShouldShowStackTrace() {
		printer.Printf("\n\n%s failed with error:\n%s", name, normalError.Error())
	} else {
		printer.Printf("\n\n%s failed with error:\n%s", name, err)
	}
}
