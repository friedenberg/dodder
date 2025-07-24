package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/lima/inventory_list_store"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/mike/env_box"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

// TODO remove and remove archive repos too
type LocalArchive struct {
	EnvRepo
}

func (cmd *LocalArchive) SetFlagSet(f *flag.FlagSet) {
}

func (cmd LocalArchive) MakeLocalArchive(
	envRepo env_repo.Env,
) repo.LocalRepo {
	repoType := envRepo.GetConfigPrivate().Blob.GetRepoType()

	switch repoType {
	case repo_type.TypeArchive:
		inventoryListBlobStore := cmd.MakeTypedInventoryListBlobStore(
			envRepo,
		)

		var inventoryListStore inventory_list_store.Store

		if err := inventoryListStore.Initialize(
			envRepo,
			nil,
			inventoryListBlobStore,
		); err != nil {
			envRepo.Cancel(err)
		}

		envBox := env_box.Make(
			envRepo,
			nil,
			nil,
		)

		inventoryListStore.SetUIDelegate(envBox.GetUIStorePrinters())

		return &inventoryListStore

	case repo_type.TypeWorkingCopy:
		return local_working_copy.MakeWithLayout(
			local_working_copy.OptionsEmpty,
			envRepo,
		)

	default:
		errors.ContextCancelWithErrorf(
			envRepo,
			"unsupported repo type: %q",
			repoType,
		)
		return nil
	}
}
