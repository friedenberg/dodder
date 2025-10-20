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
	envBlobStore := cmd.MakeEnvBlobStore(req)
	blobStoresAll := envBlobStore.GetBlobStores()

	for i, blobStore := range blobStoresAll {
		ui.Out().Printf(
			"%d: %s: %s",
			i,
			blobStore.Name,
			blobStore.GetBlobStoreDescription(),
		)
	}
}
