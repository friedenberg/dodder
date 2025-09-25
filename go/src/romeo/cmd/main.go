package cmd

import (
	"fmt"
	"os"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/stack_frame"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/quebec/commands"
)

func Run(util string) {
	utilWithExtension := extendNameIfNecessary(util)
	ctx := errors.MakeContextDefault()

	ctx.SetCancelOnSignals(
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
	)

	if err := ctx.Run(
		func(ctx interfaces.Context) {
			commands.Run(util, ctx, os.Args...)
		},
	); err != nil {
		os.Exit(handleMainErrors(ctx, utilWithExtension, err))
	}
}

func extendNameIfNecessary(name string) string {
	if name == "dodder" {
		return name
	} else {
		return fmt.Sprintf("dodder (%s)", name)
	}
}

func handleMainErrors(
	ctx interfaces.Context,
	name string,
	err error,
) (exitStatus int) {
	exitStatus = 1

	var signal errors.Signal

	if errors.As(err, &signal) {
		if signal.Signal != syscall.SIGHUP {
			ui.Err().Printf(
				"%s aborting due to signal: %s",
				name,
				signal.Signal,
			)
		}

		return exitStatus
	}

	var helpful interfaces.ErrorHelpful

	if errors.As(err, &helpful) {
		errors.PrintHelpful(ui.Err(), helpful)
		return exitStatus
	}

	if errors.Is499ClientClosedRequest(err) {
		return exitStatus
	}

	_, frames := ctx.CauseWithStackFrames()
	err = stack_frame.MakeErrorTreeOrErr(err, frames...)

	ui.CLIErrorTreeEncoder.EncodeTo(err, ui.Err())

	return exitStatus
}
