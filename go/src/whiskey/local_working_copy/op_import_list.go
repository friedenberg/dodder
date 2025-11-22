package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/tango/repo"
	"code.linenisgreat.com/dodder/go/src/uniform/remote_transfer"
)

// TODO consider moving this directly into the remote transfer package
func (local *Repo) ImportSeq(
	seq interfaces.SeqError[*sku.Transacted],
	importer repo.Importer,
) (err error) {
	local.Must(errors.MakeFuncContextFromFuncErr(local.Lock))

	var hasConflicts bool

	checkedOutPrinter := importer.GetCheckedOutPrinter()

	importer.SetCheckedOutPrinter(
		func(checkedOut *sku.CheckedOut) (err error) {
			if checkedOut.GetState() == checked_out_state.Conflicted {
				hasConflicts = true
			}

			return checkedOutPrinter(checkedOut)
		},
	)

	importErrors := errors.MakeGroupBuilder()
	missingBlobs := sku.MakeListCheckedOut()

	for object, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return err
		}

		var hasOneConflict bool

		if hasOneConflict, err = local.importOne(
			importer,
			object,
			missingBlobs,
		); err != nil {
			err = errors.Wrapf(err, "Object: %s", sku.String(object))
			importErrors.Add(err)
			err = nil
		}

		if hasOneConflict {
			hasConflicts = true
		}
	}

	checkedOutPrinter = local.GetUIStorePrinters().CheckedOut

	if missingBlobs.Len() > 0 {
		ui.Err().Printf(
			"could not import %d objects (blobs missing):\n",
			missingBlobs.Len(),
		)

		for missing := range missingBlobs.All() {
			if err = checkedOutPrinter(missing); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	if hasConflicts {
		importErrors.Add(remote_transfer.ErrNeedsMerge)
	}

	if importErrors.Len() > 0 {
		err = importErrors.GetError()
	}

	local.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))

	return err
}

func (repo *Repo) importOne(
	importur repo.Importer,
	object *sku.Transacted,
	missingBlobs *sku.HeapCheckedOut,
) (hasConflicts bool, err error) {
	var checkedOut *sku.CheckedOut
	checkedOut, err = importur.Import(object)
	defer sku.GetCheckedOutPool().Put(checkedOut)

	if err == nil {
		if checkedOut.GetState() == checked_out_state.Conflicted {
			hasConflicts = true
		}

		return hasConflicts, err
	}

	if errors.Is(err, remote_transfer.ErrSkipped) {
		err = nil
		return hasConflicts, err
	} else if errors.Is(err, collections.ErrExists) {
		err = nil
		return hasConflicts, err
	} else if genres.IsErrUnsupportedGenre(err) {
		err = nil
		return hasConflicts, err
	} else if env_dir.IsErrBlobMissing(err) {
		checkedOut := sku.GetCheckedOutPool().Get()
		sku.TransactedResetter.ResetWith(
			checkedOut.GetSkuExternal(),
			object,
		)
		checkedOut.SetState(checked_out_state.Untracked)

		missingBlobs.Add(checkedOut)

		return hasConflicts, err
	}

	return hasConflicts, err
}
