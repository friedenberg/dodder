package store_fs

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) UpdateTransacted(internal *sku.Transacted) (err error) {
	item, ok := store.Get(&internal.ObjectId)

	if !ok {
		return
	}

	var external *sku.Transacted

	if external, err = store.ReadExternalFromItem(
		sku.CommitOptions{
			StoreOptions: sku.StoreOptions{
				UpdateTai: true,
			},
		},
		item,
		internal,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	sku.Resetter.ResetWith(internal, external)
	sku.GetTransactedPool().Put(external)

	return
}

func (store *Store) ReadOneExternalObjectReader(
	reader io.Reader,
	external *sku.Transacted,
) (err error) {
	if _, err = store.metadataTextParser.ParseMetadata(
		reader,
		external,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
