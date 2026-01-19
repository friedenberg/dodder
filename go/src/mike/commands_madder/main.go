package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/config_cli"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
)

// TODO remove flags related to box format
var utility = command.MakeUtility("madder", config_cli.Default())

func GetUtility() command.Utility {
	return utility
}
