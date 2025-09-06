package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"blob_store-list",
		&BlobList{},
	)
}

type BlobList struct {
	command_components.EnvRepo
}

func (cmd *BlobList) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
}

func (cmd BlobList) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	blobStoresAll := envRepo.GetBlobStores()

	for _, blobStore := range blobStoresAll {
		ui.Out().Printf("%s", blobStore.Name)
	}
}
