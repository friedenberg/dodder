package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/juliett/object_metadata"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (store *Store) ReadExternalAndMergeIfNecessary(
	left, mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if mother == nil {
		return err
	}

	var checkedOut *sku.CheckedOut

	if checkedOut, err = store.ReadCheckedOutFromTransacted(
		options.RepoId,
		mother,
	); err != nil {
		if errors.IsNotExist(err) || collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	defer store.PutCheckedOutLike(checkedOut)

	right := checkedOut.GetSkuExternal().GetSku()

	// TODO switch to using mother
	motherEqualsExternal := object_metadata.EqualerSansTai.Equals(
		right.GetMetadata(),
		checkedOut.GetSku().GetMetadata(),
	)

	if motherEqualsExternal {
		checkoutOptions := checkout_options.OptionsWithoutMode{
			Force: true,
		}

		sku.TransactedResetter.ResetWithExceptFields(right, left)

		if err = store.UpdateCheckoutFromCheckedOut(
			checkoutOptions,
			checkedOut,
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}

	if err = right.SetMother(mother); err != nil {
		err = errors.Wrap(err)
		return err
	}

	conflicted := sku.Conflicted{
		CheckedOut: checkedOut,
		Local:      left,
		Base:       mother,
		Remote:     right,
	}

	if err = store.MergeConflicted(conflicted); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
