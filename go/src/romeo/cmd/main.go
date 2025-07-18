package cmd

import (
	"fmt"
	"os"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/quebec/commands"
)

func Run(name string) {
	name = extendNameIfNecessary(name)
	ctx := errors.MakeContextDefault()

	ctx.SetCancelOnSignals(
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
	)

	if err := ctx.Run(
		func(ctx interfaces.Context) {
			commands.Run(ctx, os.Args...)
		},
	); err != nil {
		os.Exit(handleMainErrors(ctx, name, err))
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

		return
	}

	var helpful errors.Helpful

	if errors.As(err, &helpful) {
		errors.PrintHelpful(ui.Err(), helpful)
		return
	}

	if errors.Is499ClientClosedRequest(err) {
		return
	}

	_, stackFrames := ctx.CauseWithStackFrames()

	errors.PrintStackTracerIfNecessary(ui.Err(), name, err, stackFrames)

	return
}
