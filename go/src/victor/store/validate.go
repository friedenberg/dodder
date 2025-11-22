package store

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (store *Store) validateAndFinalize(
	daughter *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if err = store.validateIfNecessary(daughter, mother, options); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = store.finalizer.WriteLockfile(
		daughter,
		store.streamIndex.ReadOneMarklIdAdded,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (store *Store) validateIfNecessary(
	daughter *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if !options.Validate {
		return err
	}

	switch daughter.GetSku().GetGenre() {
	case genres.Type:
		tipe := daughter.GetType()

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
