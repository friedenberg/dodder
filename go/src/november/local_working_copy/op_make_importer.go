package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (local *Repo) MakeImporter(
	options sku.ImporterOptions,
	storeOptions sku.StoreOptions,
) (importer sku.Importer) {
	return local.GetStore().MakeImporter(options, storeOptions)
}
