package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	command.Register(
		"blob_store-list",
		&List{},
	)
}

type List struct {
	command_components_madder.EnvRepo
}

var _ interfaces.CommandComponentWriter = (*List)(nil)

func (cmd *List) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
}

func (cmd List) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	blobStoresAll := envRepo.GetBlobStores()

	for _, blobStore := range blobStoresAll {
		ui.Out().Printf("%s", blobStore.Name)
	}
}
