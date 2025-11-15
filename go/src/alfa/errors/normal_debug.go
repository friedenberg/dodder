//go:build debug

package errors

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/_/stack_frame"
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
