package command

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/collections_slice"
)

type CommandLineInput struct {
	// TODO replace with collections_slice.String
	FlagsOrArgs          []string
	InProgress           string
	ContainsDoubleHyphen bool

	Args collections_slice.String
	Argi int

	// TODO replace with collections_slice.String
	consumed []consumedArg
}

type consumedArg struct {
	name, value string
}

func (arg consumedArg) String() string {
	if arg.name == "" {
		return fmt.Sprintf("%q", arg.value)
	} else {
		return fmt.Sprintf("%s:%q", arg.name, arg.value)
	}
}

func (commandLine CommandLineInput) LastArg() (arg string, ok bool) {
	argc := len(commandLine.FlagsOrArgs)

	if argc > 0 {
		ok = true
		arg = commandLine.FlagsOrArgs[argc-1]
	}

	return arg, ok
}

func (commandLine CommandLineInput) LastCompleteArg() (arg string, ok bool) {
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
