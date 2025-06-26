package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (s *Store) ReadExternalAndMergeIfNecessary(
	left, parent *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	if parent == nil {
		return
	}

	var co *sku.CheckedOut

	if co, err = s.ReadCheckedOutFromTransacted(
		options.RepoId,
		parent,
	); err != nil {
		err = nil
		return
	}

	defer s.PutCheckedOutLike(co)

	right := co.GetSkuExternal().GetSku()

	parentEqualsExternal := right.Metadata.EqualsSansTai(&co.GetSku().Metadata)

	if parentEqualsExternal {
		op := checkout_options.OptionsWithoutMode{
			Force: true,
		}

		sku.TransactedResetter.ResetWithExceptFields(right, left)

		if err = s.UpdateCheckoutFromCheckedOut(
			op,
			co,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	conflicted := sku.Conflicted{
		CheckedOut: co,
		Local:      left,
		Base:       parent,
		Remote:     right,
	}

	if err = s.MergeConflicted(conflicted); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
