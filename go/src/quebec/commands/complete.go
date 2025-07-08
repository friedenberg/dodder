package commands

import (
	"flag"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"complete",
		&Complete{},
	)
}

type Complete struct {
	command_components.Env
	command_components.Complete

	bashStyle  bool
	inProgress string
}

func (cmd Complete) GetDescription() command.Description {
	return command.Description{
		Short: "complete a command-line",
	}
}

func (cmd *Complete) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(&cmd.bashStyle, "bash-style", false, "")
	f.StringVar(&cmd.inProgress, "in-progress", "", "")
}

func (cmd Complete) Run(req command.Request) {
	cmds := command.Commands()
	envLocal := cmd.MakeEnv(req)

	// TODO extract into constructor
	// TODO find double-hyphen
	// TODO keep track of all args
	commandLine := command.CommandLine{
		FlagsOrArgs: req.PeekArgs(),
		InProgress:  cmd.inProgress,
	}

	// TODO determine state:
	// bare: `dodder`
	// subcommand or arg or flag:
	//  - `dodder subcommand`
	//  - `dodder subcommand -flag=true`
	//  - `dodder subcommand -flag value`
	// flag: `dodder subcommand -flag`
	lastArg, hasLastArg := commandLine.LastArg()

	if !hasLastArg {
		cmd.completeSubcommands(envLocal, commandLine, cmds)
		return
	}

	name := req.PopArg("name")
	subcmd, foundSubcmd := cmds[name]

	if !foundSubcmd {
		cmd.completeSubcommands(envLocal, commandLine, cmds)
		return
	}

	flagSet := flag.NewFlagSet(name, flag.ContinueOnError)
	flagSet.SetOutput(io.Discard)
	(&repo_config_cli.Config{}).SetFlagSet(flagSet)
	subcmd.SetFlagSet(flagSet)

	var containsDoubleHyphen bool

	for _, arg := range commandLine.FlagsOrArgs {
		if arg == "--" {
			containsDoubleHyphen = true
			break
		}
	}

	if !containsDoubleHyphen &&
		cmd.completeSubcommandFlags(
			req,
			envLocal,
			subcmd,
			flagSet,
			commandLine,
			lastArg,
		) {
		return
	}

	cmd.completeSubcommandArgs(req, envLocal, subcmd, commandLine)
}

func (cmd Complete) completeSubcommands(
	envLocal env_local.Env,
	commandLine command.CommandLine,
	cmds map[string]command.Command,
) {
	for name, subcmd := range cmds {
		cmd.completeSubcommand(envLocal, name, subcmd)
	}
}

func (cmd Complete) completeSubcommand(
	envLocal env_local.Env,
	name string,
	subcmd command.Command,
) {
	var shortDescription string

	if hasDescription, ok := subcmd.(command.HasDescription); ok {
		description := hasDescription.GetDescription()
		shortDescription = description.Short
	}

	if shortDescription != "" {
		envLocal.GetUI().Printf("%s\t%s", name, shortDescription)
	} else {
		envLocal.GetUI().Printf("%s", name)
	}
}

func (cmd Complete) completeSubcommandArgs(
	req command.Request,
	envLocal env_local.Env,
	subcmd command.Command,
	commandLine command.CommandLine,
) {
	if subcmd == nil {
		return
	}

	completer, isCompleter := subcmd.(command.Completer)

	if !isCompleter {
		return
	}

	completer.Complete(req, envLocal, commandLine)
}

func (cmd Complete) completeSubcommandFlags(
	req command.Request,
	envLocal env_local.Env,
	subcmd command.Command,
	flagSet *flag.FlagSet,
	commandLine command.CommandLine,
	lastArg string,
) (shouldNotCompleteArgs bool) {
	if subcmd == nil {
		return
	}

	if strings.HasPrefix(lastArg, "-") && commandLine.InProgress != "" {
		shouldNotCompleteArgs = true
	} else if commandLine.InProgress != "" && len(commandLine.FlagsOrArgs) > 1 {
		lastArg = commandLine.FlagsOrArgs[len(commandLine.FlagsOrArgs)-2]
		commandLine.InProgress = ""
		shouldNotCompleteArgs = strings.HasPrefix(lastArg, "-")
	}

	if commandLine.InProgress != "" {
		flagSet.VisitAll(func(flag *flag.Flag) {
			envLocal.GetUI().Printf("-%s\t%s", flag.Name, flag.Usage)
		})
	} else if err := flagSet.Parse([]string{lastArg}); err != nil {
		cmd.completeSubcommandFlagOnParseError(
			req,
			envLocal,
			subcmd,
			flagSet,
			commandLine,
			err,
		)
	} else {
		flagSet.VisitAll(func(flag *flag.Flag) {
			envLocal.GetUI().Printf("-%s\t%s", flag.Name, flag.Usage)
		})
	}

	return
}

func (cmd Complete) completeSubcommandFlagOnParseError(
	req command.Request,
	envLocal env_local.Env,
	subcmd command.Command,
	flagSet *flag.FlagSet,
	commandLine command.CommandLine,
	err error,
) {
	if subcmd == nil {
		return
	}

	after, found := strings.CutPrefix(
		err.Error(),
		"flag needs an argument: -",
	)

	if !found {
		envLocal.CancelWithBadRequestf(err.Error())
		return
	}

	var flagCompleter command.Completer

	var flag *flag.Flag

	if flag = flagSet.Lookup(after); flag != nil {
		flagCompleter, _ = flag.Value.(command.Completer)
	}

	if flagCompleter != nil {
		flagCompleter.Complete(req, envLocal, commandLine)
		return
	}

	req.CancelWithBadRequestf("no completion available for flag: %q, %#v", after, flag)
}
