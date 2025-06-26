package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type RemoteTransfer struct {
	Remote
	repo.RemoteTransferOptions
}

func (cmd *RemoteTransfer) SetFlagSet(f *flag.FlagSet) {
	cmd.Remote.SetFlagSet(f)
	cmd.RemoteTransferOptions.SetFlagSet(f)
}

func (cmd *RemoteTransfer) PushAllToArchive(
	req command.Request,
	local, remote repo.Repo,
) {
	remoteInventoryListStore := remote.GetInventoryListStore()
	localInventoryListStore := local.GetInventoryListStore()

	// TODO fetch tais of inventory lists we've pushed

	for list, err := range localInventoryListStore.IterAllInventoryLists() {
		// TODO continue to next if we pushed this list already

		if err != nil {
			req.CancelWithError(err)
			return
		}

		if err := remoteInventoryListStore.ImportInventoryList(
			local.GetBlobStore(),
			list,
		); err != nil {
			req.CancelWithError(err)
			return
		}

		// TODO add this list's tai to the lists we've pushed so far
	}

	// TODO persist all the tais of lists we've pushed so far to a cache
}
