package command

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/cli"
	"code.linenisgreat.com/dodder/go/src/juliett/env_local"
)

type SupportsCompletion interface {
	SupportsCompletion()
}

type CLICompleter = cli.CLICompleter

type CommandLine struct {
	FlagsOrArgs []string
	InProgress  string
}

func (commandLine CommandLine) LastArg() (arg string, ok bool) {
	argc := len(commandLine.FlagsOrArgs)

	if argc > 0 {
		ok = true
		arg = commandLine.FlagsOrArgs[argc-1]
	}

	return arg, ok
}

func (commandLine CommandLine) LastCompleteArg() (arg string, ok bool) {
	argc := len(commandLine.FlagsOrArgs)

	if commandLine.InProgress != "" {
		argc -= 1
	}

	if argc > 0 {
		ok = true
		arg = commandLine.FlagsOrArgs[argc-1]
	}

	return arg, ok
}

type Completion struct {
	Value, Description string
}

type Completer interface {
	Complete(Request, env_local.Env, CommandLine)
}

type FuncCompleter func(Request, env_local.Env, CommandLine)

type FlagValueCompleter struct {
	interfaces.FlagValue
	FuncCompleter
}

func (completer FlagValueCompleter) String() string {
	// TODO still not sure why this condition can exist, but this makes the
	// output
	// nice
	if completer.FlagValue == nil {
		return ""
	} else {
		return completer.FlagValue.String()
	}
}

func (completer FlagValueCompleter) Complete(
	req Request,
	envLocal env_local.Env,
	commandLine CommandLine,
) {
	completer.FuncCompleter(req, envLocal, commandLine)
}
