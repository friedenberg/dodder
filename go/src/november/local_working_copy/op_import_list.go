package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/importer"
)

func (local *Repo) ImportSeq(
	seq interfaces.SeqError[*sku.Transacted],
	importerr sku.Importer,
) (err error) {
	local.Must(errors.MakeFuncContextFromFuncErr(local.Lock))

	var hasConflicts bool

	checkedOutPrinter := importerr.GetCheckedOutPrinter()

	importerr.SetCheckedOutPrinter(
		func(co *sku.CheckedOut) (err error) {
			if co.GetState() == checked_out_state.Conflicted {
				hasConflicts = true
			}

			return checkedOutPrinter(co)
		},
	)

	importErrors := errors.MakeGroupBuilder()
	missingBlobs := sku.MakeListCheckedOut()

	for object, iterErr := range seq {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		checkedOut, importError := importerr.Import(object)

		func() {
			defer sku.GetCheckedOutPool().Put(checkedOut)

			if importError == nil {
				if checkedOut.GetState() == checked_out_state.Conflicted {
					hasConflicts = true
				}

				return
			}

			if errors.Is(importError, collections.ErrExists) {
				return
			}

			if genres.IsErrUnsupportedGenre(importError) {
				return
			}

			if env_dir.IsErrBlobMissing(importError) {
				checkedOut := sku.GetCheckedOutPool().Get()
				sku.TransactedResetter.ResetWith(
					checkedOut.GetSkuExternal(),
					object,
				)
				checkedOut.SetState(checked_out_state.Untracked)

				missingBlobs.Add(checkedOut)

				return
			}

			importErrors.Add(errors.Wrapf(err, "Sku: %s", sku.String(object)))
		}()
	}

	checkedOutPrinter = local.GetUIStorePrinters().CheckedOutCheckedOut

	if missingBlobs.Len() > 0 {
		ui.Err().Printf(
			"could not import %d objects (blobs missing):\n",
			missingBlobs.Len(),
		)

		for missing := range missingBlobs.All() {
			if err = checkedOutPrinter(missing); err != nil {
				err = errors.Wrap(err)
				return
			}
		}
	}

	if hasConflicts {
		importErrors.Add(importer.ErrNeedsMerge)
	}

	if importErrors.Len() > 0 {
		err = importErrors
	}

	local.Must(errors.MakeFuncContextFromFuncErr(local.Unlock))

	return
}
