package inventory_list_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) MakeImporter(
	options sku.ImporterOptions,
	storeOptions sku.StoreOptions,
) sku.Importer {
	panic(errors.Err405MethodNotAllowed)
}

func (store *Store) ImportSeq(
	seq sku.Seq,
	importer sku.Importer,
) (err error) {
	return errors.Err405MethodNotAllowed
}

// TODO split into public and private parts, where public includes writing the
// skus AND the list, while private writes just the list
func (store *Store) ImportInventoryList(
	remoteBlobStore interfaces.BlobStore,
	remoteListObject *sku.Transacted,
) (err error) {
	return errors.Err405MethodNotAllowed
}
