package command

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

type CommandComponentReader interface {
	GetCLIFlags() []string
}

type CommandComponent interface {
	CommandComponentReader
	interfaces.CommandComponentWriter
}
