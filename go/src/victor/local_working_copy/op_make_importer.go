package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/tango/repo"
	"code.linenisgreat.com/dodder/go/src/uniform/remote_transfer"
)

func (local *Repo) MakeImporter(
	options repo.ImporterOptions,
	storeOptions sku.StoreOptions,
) repo.Importer {
	store := local.GetStore()

	return remote_transfer.Make(
		options,
		storeOptions,
		store.GetEnvRepo(),
		store.GetTypedBlobStore().InventoryList,
		store.GetStreamIndex(),
		local.GetEnvWorkspace().GetStore(),
		store,
	)
}
