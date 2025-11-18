package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/november/commands_madder"
)

var utility = command.MakeUtility(
	"dodder",
).MergeUtilityWithPrefix(
	commands_madder.GetUtility(),
	"blob_store",
)

func GetUtility(name string) command.Utility {
	return utility
}
