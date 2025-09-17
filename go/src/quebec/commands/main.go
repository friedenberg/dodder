package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	_ "code.linenisgreat.com/dodder/go/src/lima/commands_madder"
)

// TODO switch to commands.Request
func Run(util string, ctx interfaces.Context, args ...string) {
	if len(args) <= 1 {
		PrintUsage(
			util,
			ctx,
			errors.BadRequestf("No subcommand provided."),
		)
	}

	cmds := command.CommandsFor(util)

	var cmd command.Command
	var ok bool

	name := args[1]

	if cmd, ok = cmds[name]; !ok {
		PrintUsage(
			util,
			ctx,
			errors.BadRequestf("No subcommand %q", name),
		)
	}

	flagSet := flags.NewFlagSet(name, flags.ContinueOnError)

	if cmd, ok := cmd.(interfaces.CommandComponentWriter); ok {
		cmd.SetFlagDefinitions(flagSet)
	}

	args = args[2:]

	configCli := repo_config_cli.Default()
	configCli.SetFlagDefinitions(flagSet)

	if err := flagSet.Parse(args); err != nil {
		ctx.Cancel(err)
	}

	req := command.MakeRequest(
		util,
		ctx,
		configCli,
		flagSet,
	)

	cmd.Run(req)
}
