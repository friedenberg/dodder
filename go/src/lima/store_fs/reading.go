package store_fs

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (s *Store) UpdateTransacted(internal *sku.Transacted) (err error) {
	item, ok := s.Get(&internal.ObjectId)

	if !ok {
		return
	}

	var external *sku.Transacted

	if external, err = s.ReadExternalFromItem(
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

func (s *Store) ReadOneExternalObjectReader(
	r io.Reader,
	external *sku.Transacted,
) (err error) {
	if _, err = s.metadataTextParser.ParseMetadata(r, external); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
