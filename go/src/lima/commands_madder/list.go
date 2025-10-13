package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("list", &List{})
}

type List struct {
	command_components_madder.EnvBlobStore
}

var _ interfaces.CommandComponentWriter = (*List)(nil)

func (cmd *List) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
}

func (cmd List) Run(req command.Request) {
	envRepo := cmd.MakeEnvBlobStore(req)
	blobStoresAll := envRepo.GetBlobStores()

	for _, blobStore := range blobStoresAll {
		ui.Out().Printf("%#v", blobStore)
	}
}
