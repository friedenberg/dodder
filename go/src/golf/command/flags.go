package command

import "code.linenisgreat.com/dodder/go/src/bravo/flags"

type CommandComponentReader interface {
	GetCLIFlags() []string
}

type CommandComponent interface {
	CommandComponentReader
	flags.CommandComponentWriter
}
