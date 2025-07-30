//go:build debug

package errors

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

// TODO move to UI
func PrintCommandFailureWithStackTracerIfNecessary(
	printer interfaces.Printer,
	name string,
	err error,
	stackFrames []stack_frame.Frame,
) {
	PrintWithStackFramesIfNecessary(
		printer,
		fmt.Sprintf("%s failed with error: %s", name, err),
		stackFrames,
	)
}

func PrintWithStackFramesIfNecessary(
	printer interfaces.Printer,
	message string,
	stackFrames []stack_frame.Frame,
) {
	if len(stackFrames) > 0 && DebugBuild {
		printer.Printf("\n\n%s\n", stackFrames, message)
	} else {
		printer.Printf("\n\n%s", message)
	}
}
