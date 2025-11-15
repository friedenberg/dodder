//go:build debug

package errors

import (
	"code.linenisgreat.com/dodder/go/src/_/stack_frame"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func PrintWithStackFramesIfNecessary(
	printer interfaces.Printer,
	message string,
	stackFrames []stack_frame.Frame,
) {
	if len(stackFrames) > 0 && debugBuild {
		printer.Printf("\n\n%s\n", stackFrames, message)
	} else {
		printer.Printf("\n\n%s", message)
	}
}
