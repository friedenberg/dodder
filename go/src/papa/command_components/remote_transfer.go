package command_components

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type RemoteTransfer struct {
	Remote
	repo.RemoteTransferOptions
}

func (cmd *RemoteTransfer) SetFlagSet(f *flags.FlagSet) {
	cmd.Remote.SetFlagSet(f)
	cmd.RemoteTransferOptions.SetFlagSet(f)
}

func (cmd *RemoteTransfer) PushAllToArchive(
	req command.Request,
	local, remote repo.Repo,
) {
	req.Cancel(errors.Err405MethodNotAllowed)
	//TODO use the ideas from below for pushes
	// remoteInventoryListStore := remote.GetInventoryListStore()
	// localInventoryListStore := local.GetInventoryListStore()

	// // TODO fetch tais of inventory lists we've pushed

	// for list, err := range localInventoryListStore.IterAllInventoryLists() {
	// 	// TODO continue to next if we pushed this list already

	// 	if err != nil {
	// 		req.Cancel(err)
	// 		return
	// 	}

	// 	if err := remoteInventoryListStore.ImportInventoryList(
	// 		local.GetBlobStore(),
	// 		list,
	// 	); err != nil {
	// 		req.Cancel(err)
	// 		return
	// 	}

	// 	// TODO add this list's tai to the lists we've pushed so far
	// }

	// // TODO persist all the tais of lists we've pushed so far to a cache
}
