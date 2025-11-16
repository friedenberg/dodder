package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) ReadExternalAndMergeIfNecessary(
	left, mother *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if mother == nil {
		return err
	}

	var co *sku.CheckedOut

	if co, err = store.ReadCheckedOutFromTransacted(
		options.RepoId,
		mother,
	); err != nil {
		err = nil
		return err
	}

	defer store.PutCheckedOutLike(co)

	right := co.GetSkuExternal().GetSku()

	// TODO switch to using mother
	motherEqualsExternal := object_metadata.EqualerSansTai.Equals(right.GetMetadata(), co.GetSku().GetMetadata())

	if motherEqualsExternal {
		op := checkout_options.OptionsWithoutMode{
			Force: true,
		}

		sku.TransactedResetter.ResetWithExceptFields(right, left)

		if err = store.UpdateCheckoutFromCheckedOut(
			op,
			co,
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
		CheckedOut: co,
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
