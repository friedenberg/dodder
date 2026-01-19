package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/mike/commands_madder"
)

var utility = command.MakeUtility(
	"dodder",
	repo_config_cli.Default(),
).MergeUtilityWithPrefix(
	commands_madder.GetUtility(),
	"blob_store",
)

func GetUtility(name string) command.Utility {
	return utility
}
