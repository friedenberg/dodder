package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) validate(
	el sku.ExternalLike, mutter *sku.Transacted,
	o sku.CommitOptions,
) (err error) {
	if !o.Validate {
		return
	}

	switch el.GetSku().GetGenre() {
	case genres.Type:
		tipe := el.GetSku().GetType()

		var commonBlob type_blobs.Blob

		if commonBlob, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
			tipe,
			el.GetSku().GetBlobId(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer store.GetTypedBlobStore().Type.PutTypedBlob(tipe, commonBlob)
	}

	return
}
