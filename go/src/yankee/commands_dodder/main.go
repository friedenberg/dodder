package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/lima/commands_madder"
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
