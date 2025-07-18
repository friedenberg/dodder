//go:build debug

package errors

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
)

// TODO move to UI
func PrintStackTracerIfNecessary(
	printer interfaces.Printer,
	name string,
	err error,
	stackFrames []stack_frame.Frame,
) {
	if len(stackFrames) > 0 && DebugBuild {
		printer.Printf(
			"\n\n%s\n%s failed with error:\n%s",
			stackFrames,
			name,
			err,
		)
	} else {
		printer.Printf("\n\n%s failed with error:\n%s", name, err)
	}
}
