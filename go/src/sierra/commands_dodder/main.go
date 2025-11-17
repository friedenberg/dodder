package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/lima/commands_madder"
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
