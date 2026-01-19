package commands_madder

import "code.linenisgreat.com/dodder/go/src/kilo/command"

// TODO remove flags related to box format
var utility = command.MakeUtility("madder")

func GetUtility() command.Utility {
	return utility
}
