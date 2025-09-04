package importer

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

var ErrNeedsMerge = errors.Err409Conflict.ErrorHiddenf(
	"import failed with conflicts, merging required",
)

func Make(
	options sku.ImporterOptions,
	storeOptions sku.StoreOptions,
	envRepo env_repo.Env,
	typedInventoryListBlobStore inventory_list_coders.Closet,
	indexObject sku.Index,
	storeExternalMergeCheckedOut store_workspace.MergeCheckedOut,
	storeObject sku.RepoStore,
) sku.Importer {
	if options.BlobGenres.IsEmpty() {
		options.BlobGenres = ids.MakeGenreAll()
	}

	importer := &importer{
		typedInventoryListBlobStore: typedInventoryListBlobStore,
		index:                       indexObject,
		storeExternal:               storeExternalMergeCheckedOut,
		storeObject:                 storeObject,
		envRepo:                     envRepo,
		blobGenres:                  options.BlobGenres,
		excludeObjects:              options.ExcludeObjects,
		remoteBlobStore:             options.RemoteBlobStore,
		blobCopierDelegate:          options.BlobCopierDelegate,
		allowMergeConflicts:         options.AllowMergeConflicts,
		parentNegotiator:            options.ParentNegotiator,
		checkedOutPrinter:           options.CheckedOutPrinter,
		storeOptions:                storeOptions,
	}

	importer.committer.initialize(options, envRepo, storeObject)

	if importer.blobCopierDelegate == nil &&
		importer.remoteBlobStore != nil &&
		options.PrintCopies {
		importer.blobCopierDelegate = sku.MakeBlobCopierDelegate(
			envRepo.GetUI(),
		)
	}

	return importer
}

type importer struct {
	committer committer

	typedInventoryListBlobStore inventory_list_coders.Closet
	index                       sku.Index
	storeExternal               store_workspace.MergeCheckedOut
	storeObject                 sku.RepoStore
	envRepo                     env_repo.Env
	blobGenres                  ids.Genre
	excludeObjects              bool
	remoteBlobStore             interfaces.BlobStore
	blobCopierDelegate          interfaces.FuncIter[sku.BlobCopyResult]
	storeOptions                sku.StoreOptions
	allowMergeConflicts         bool
	parentNegotiator            sku.ParentNegotiator
	checkedOutPrinter           interfaces.FuncIter[*sku.CheckedOut]
}

func (importer importer) GetCheckedOutPrinter() interfaces.FuncIter[*sku.CheckedOut] {
	return importer.checkedOutPrinter
}

func (importer *importer) SetCheckedOutPrinter(
	printer interfaces.FuncIter[*sku.CheckedOut],
) {
	importer.checkedOutPrinter = printer
}

func (importer importer) Import(
	external *sku.Transacted,
) (checkedOut *sku.CheckedOut, err error) {
	errors.ContextContinueOrPanic(importer.envRepo)

	if err = importer.ImportBlobIfNecessary(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	if external.GetGenre() == genres.InventoryList {
		if checkedOut, err = importer.importInventoryList(external); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if checkedOut, err = importer.importLeaf(external); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (importer importer) importInventoryList(
	list *sku.Transacted,
) (checkedOut *sku.CheckedOut, err error) {
	if err = genres.InventoryList.AssertGenre(list.GetGenre()); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobDigest := list.GetBlobDigest()

	if !importer.envRepo.GetDefaultBlobStore().HasBlob(blobDigest) {
		err = env_dir.ErrBlobMissing{
			BlobId: blobDigest,
		}

		return
	}

	seq := importer.typedInventoryListBlobStore.StreamInventoryListBlobSkus(
		list,
	)

	for object, errIter := range seq {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return
		}

		if _, err = importer.Import(
			object,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	// TODO decide whether we should rewrite the imported inventory list
	// according to this repo's inventory list type
	// inventoryListTypeString :=
	// importer.envRepo.GetConfigPublic().Blob.GetInventoryListTypeString()

	// if listObject.GetType().String() != inventoryListTypeString {
	// 	listObject.Metadata.Type = ids.GetOrPanic(inventoryListTypeString).Type
	// }

	if checkedOut, err = importer.importLeaf(
		list,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (importer importer) importLeaf(
	external *sku.Transacted,
) (checkedOut *sku.CheckedOut, err error) {
	if importer.excludeObjects {
		err = ErrSkipped
		return
	}

	// TODO address this terrible hack? How should config objects be handled by
	// remotes?
	if external.GetGenre() == genres.Config {
		err = genres.MakeErrUnsupportedGenre(external.GetGenre())
		return
	}

	checkedOut = sku.GetCheckedOutPool().Get()

	sku.Resetter.ResetWith(checkedOut.GetSkuExternal(), external)

	checkedOut.GetSkuExternal().Metadata.GetObjectDigestMutable().Reset()

	// TODO confirm repo pub key

	// TODO set this as an importer option
	if checkedOut.GetSkuExternal().Metadata.GetObjectSig().IsNull() {
		if err = checkedOut.GetSkuExternal().FinalizeAndSignOverwrite(
			importer.envRepo.GetConfigPrivate().Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = checkedOut.GetSkuExternal().FinalizeUsingObject(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if importer.index != nil {
		_, err = importer.index.ReadOneObjectIdTai(
			checkedOut.GetSkuExternal().GetObjectId(),
			checkedOut.GetSkuExternal().GetTai(),
		)

		if err == nil {
			err = collections.ErrExists
			return
		} else if collections.IsErrNotFound(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	ui.TodoP4("cleanup")
	if err = importer.storeObject.ReadOneInto(
		checkedOut.GetSkuExternal().GetObjectId(),
		checkedOut.GetSku(),
	); err != nil {
		if collections.IsErrNotFound(err) {
			if err = importer.importNewObject(
				checkedOut.GetSkuExternal(),
			); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		} else {
			err = errors.Wrapf(err, "ObjectId: %s", external.GetObjectId())
		}

		return
	}

	var commitOptions sku.CommitOptions

	// TODO extra commit option setting into its own function
	if importer.storeExternal != nil {
		if commitOptions, err = importer.storeExternal.MergeCheckedOut(
			checkedOut,
			importer.parentNegotiator,
			importer.allowMergeConflicts,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if checkedOut.GetState() == checked_out_state.Conflicted {
			if !importer.allowMergeConflicts {
				if err = importer.checkedOutPrinter(checkedOut); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		}
	}

	commitOptions.Validate = false

	if err = importer.committer.Commit(
		checkedOut.GetSkuExternal(),
		commitOptions,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = importer.checkedOutPrinter(checkedOut); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (importer importer) importNewObject(
	object *sku.Transacted,
) (err error) {
	options := sku.CommitOptions{
		Clock:              object,
		StoreOptions:       importer.storeOptions,
		DontAddMissingType: true,
	}

	options.UpdateTai = false

	if err = importer.committer.Commit(
		object,
		options,
	); err != nil {
		err = errors.WrapExceptSentinel(err, collections.ErrExists)
		return
	}

	return
}

func (importer importer) ImportBlobIfNecessary(
	object *sku.Transacted,
) (err error) {
	blobId := object.GetBlobDigest()

	if importer.remoteBlobStore == nil {
		// when this is a dumb HTTP remote, we expect local to push the missing
		// objects to us after the import call

		n := int64(-1)

		if importer.envRepo.GetDefaultBlobStore().HasBlob(blobId) {
			n = -2
		}

		if importer.blobCopierDelegate != nil {
			if err = importer.blobCopierDelegate(
				sku.BlobCopyResult{
					Transacted: object,
					MarklId:    blobId,
					N:          n,
				},
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		return
	}

	if !importer.blobGenres.Contains(object.GetGenre()) {
		return
	}

	var progressWriter env_ui.ProgressWriter

	if err = errors.RunChildContextWithPrintTicker(
		importer.envRepo,
		func(ctx interfaces.Context) {
			var n int64

			if n, err = blob_stores.CopyBlobIfNecessary(
				importer.envRepo,
				importer.envRepo.GetDefaultBlobStore(),
				importer.remoteBlobStore,
				blobId,
				&progressWriter,
			); err != nil {
				if errors.Is(err, &env_dir.ErrBlobAlreadyExists{}) {
					err = nil
				} else {
					// TODO add context that this could not be copied from the
					// remote blob
					// store
					err = errors.Wrap(err)
					return
				}

				return
			}

			if importer.blobCopierDelegate != nil {
				if err = importer.blobCopierDelegate(
					sku.BlobCopyResult{
						Transacted: object,
						MarklId:    blobId,
						N:          n,
					},
				); err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		},
		func(time time.Time) {
			ui.Err().Printf(
				"Copying %s... (%s written)",
				blobId,
				progressWriter.GetWrittenHumanString(),
			)
		},
		3*time.Second,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
