package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
)

// TODO switch to config_cli instead
var utility = command.MakeUtility("madder", repo_config_cli.Default())

func GetUtility() command.Utility {
	return utility
}
