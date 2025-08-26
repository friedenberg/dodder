package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) validate(
	object sku.ExternalLike,
	motherObject *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if !options.Validate {
		return
	}

	switch object.GetSku().GetGenre() {
	case genres.Type:
		tipe := object.GetSku().GetType()

		var repool interfaces.FuncRepool

		if _, repool, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
			tipe,
			object.GetSku().GetBlobDigest(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer repool()
	}

	return
}
