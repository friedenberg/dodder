package inventory_list_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_importer "code.linenisgreat.com/dodder/go/src/mike/importer"
)

func (store *Store) MakeImporter(
	options sku.ImporterOptions,
	storeOptions sku.StoreOptions,
) sku.Importer {
	importer := pkg_importer.Make(
		options,
		storeOptions,
		store.envRepo,
		store.GetInventoryListCoderCloset(),
		nil,
		nil,
		store,
	)

	return importer
}

func (store *Store) ImportSeq(
	seq sku.Seq,
	importer sku.Importer,
) (err error) {
	var hasConflicts bool

	checkedOutPrinter := importer.GetCheckedOutPrinter()

	importer.SetCheckedOutPrinter(
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

		checkedOut, importError := importer.Import(object)

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

	checkedOutPrinter = store.ui.CheckedOutCheckedOut

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
		importErrors.Add(pkg_importer.ErrNeedsMerge)
	}

	if importErrors.Len() > 0 {
		err = importErrors
	}

	return
}

// TODO split into public and private parts, where public includes writing the
// skus AND the list, while private writes just the list
func (store *Store) ImportInventoryList(
	remoteBlobStore interfaces.BlobStore,
	remoteListObject *sku.Transacted,
) (err error) {
	if !remoteListObject.GetType().Equals(
		ids.GetOrPanic(
			store.envRepo.GetConfigPublic().Blob.GetInventoryListTypeString(),
		).Type) {
		return store.importInventoryListAndConvert(
			remoteBlobStore,
			remoteListObject,
		)
	}

	var blobReader interfaces.ReadCloseBlobIdGetter

	if blobReader, err = remoteBlobStore.BlobReader(
		remoteListObject.GetBlobId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(blobReader)
	defer repoolBufferedReader()

	var list *sku.List

	inventoryListCoderCloset := store.GetInventoryListCoderCloset()

	if list, err = inventoryListCoderCloset.ReadInventoryListBlob(
		store.GetEnvRepo(),
		remoteListObject.GetType(),
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for childObject := range list.All() {
		if err = childObject.CalculateObjectDigests(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if _, err = blob_stores.CopyBlobIfNecessary(
			store.GetEnvRepo().GetEnv(),
			store.blobBlobStore,
			remoteBlobStore,
			childObject.GetBlobId(),
			nil,
		); err != nil {
			if env_dir.IsErrBlobAlreadyExists(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}

			continue
		}
	}

	// TODO transform listObject into local type if necessary

	if err = store.WriteInventoryListBlob(
		remoteBlobStore,
		remoteListObject,
		list,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.WriteInventoryListObject(
		remoteListObject,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) importInventoryListAndConvert(
	remoteBlobStore interfaces.BlobStore,
	remoteListObject *sku.Transacted,
) (err error) {
	var blobReader interfaces.ReadCloseBlobIdGetter

	if blobReader, err = remoteBlobStore.BlobReader(
		remoteListObject.GetBlobId(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(blobReader)
	defer repoolBufferedReader()

	var list *sku.List

	inventoryListCoderCloset := store.GetInventoryListCoderCloset()

	if list, err = inventoryListCoderCloset.ReadInventoryListBlob(
		store.GetEnvRepo(),
		remoteListObject.GetType(),
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for childObject := range list.All() {
		if err = childObject.CalculateObjectDigests(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if _, err = blob_stores.CopyBlobIfNecessary(
			store.GetEnvRepo().GetEnv(),
			store.blobBlobStore,
			remoteBlobStore,
			childObject.GetBlobId(),
			nil,
		); err != nil {
			if env_dir.IsErrBlobAlreadyExists(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}

			continue
		}
	}

	localListObject := remoteListObject.CloneTransacted()
	defer sku.GetTransactedPool().Put(localListObject)

	localListObject.Metadata.Type = ids.GetOrPanic(
		store.envRepo.GetConfigPublic().Blob.GetInventoryListTypeString(),
	).Type

	if err = store.WriteInventoryListBlob(
		remoteBlobStore,
		localListObject,
		list,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.WriteInventoryListObject(
		remoteListObject,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
