package command

type Args struct {
	args []string
	argi int

	consumed []consumedArg
}

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
