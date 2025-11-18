package commands_madder

import "code.linenisgreat.com/dodder/go/src/kilo/command"

var utility = command.MakeUtility("madder")

func GetUtility() command.Utility {
	return utility
}
