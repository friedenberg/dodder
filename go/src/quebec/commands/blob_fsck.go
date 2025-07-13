package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"blob-fsck",
		&BlobFsck{},
	)
}

type BlobFsck struct {
	command_components.EnvRepo
}

func (cmd *BlobFsck) SetFlagSet(flagSet *flag.FlagSet) {
}

func (cmd BlobFsck) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	blobStores := envRepo.GetBlobStores()

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	for i, blobStore := range blobStores {
		ui.Out().Printf("%d: %s", i, blobStore.GetBlobStoreDescription())
	}
}
