package cmd

import (
	"os"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/quebec/commands"
)

func Run(name string) {
	ctx := errors.MakeContextDefault()

	ctx.SetCancelOnSignals(
		syscall.SIGTERM,
		syscall.SIGINT,
		syscall.SIGHUP,
	)

	if err := ctx.Run(
		func(ctx errors.Context) {
			commands.Run(ctx, os.Args...)
		},
	); err != nil {
		os.Exit(handleMainErrors(ctx, name, err))
	}
}

func handleMainErrors(ctx errors.Context, name string, err error) (exitStatus int) {
	exitStatus = 1

	var signal errors.Signal

	if errors.As(err, &signal) {
		if signal.Signal != syscall.SIGHUP {
			ui.Err().Print("dodder (%s) aborting due to signal: %s", name, signal.Signal)
		}

		return
	}

	var helpful errors.Helpful

	if errors.As(err, &helpful) {
		errors.PrintHelpful(ui.Err(), helpful)
		return
	}

	var normalError errors.StackTracer

	if errors.As(err, &normalError) && !normalError.ShouldShowStackTrace() {
		ui.Err().Printf("\n\ndodder (%s) failed with error:\n%s", name, normalError.Error())
	} else {
		ui.Err().Printf("\n\ndodder (%s) failed with error:\n%s", name, err)
	}

	return
}
