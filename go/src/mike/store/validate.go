package store

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) validate(
	daughter sku.ExternalLike,
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if !options.Validate {
		return err
	}

	switch daughter.GetSku().GetGenre() {
	case genres.Type:
		tipe := daughter.GetSku().GetType()

		var repool interfaces.FuncRepool

		if _, repool, _, err = store.GetTypedBlobStore().Type.ParseTypedBlob(
			tipe,
			daughter.GetSku().GetBlobDigest(),
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		defer repool()
	}

	return err
}
