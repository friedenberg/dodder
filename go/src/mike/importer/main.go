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
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

var ErrNeedsMerge = errors.NewNormal("needs merge")

type ImporterOptions = sku.ImporterOptions

func Make(
	options ImporterOptions,
	storeOptions sku.StoreOptions,
	envRepo env_repo.Env,
	typedInventoryListBlobStore typed_blob_store.InventoryList,
	indexObject sku.IndexObject,
	storeExternalMergeCheckedOut store_workspace.MergeCheckedOut,
	storeObject sku.RepoStore,
) sku.Importer {
	if options.BlobGenres.IsEmpty() {
		options.BlobGenres = ids.MakeGenreAll()
	}

	importer := &importer{
		typedInventoryListBlobStore: typedInventoryListBlobStore,
		indexObject:                 indexObject,
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
	typedInventoryListBlobStore typed_blob_store.InventoryList
	indexObject                 sku.IndexObject
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
	p interfaces.FuncIter[*sku.CheckedOut],
) {
	importer.checkedOutPrinter = p
}

func (importer importer) Import(
	external *sku.Transacted,
) (co *sku.CheckedOut, err error) {
	errors.ContextContinueOrPanic(importer.envRepo)

	if err = importer.ImportBlobIfNecessary(external); err != nil {
		err = errors.Wrap(err)
		return
	}

	if external.GetGenre() == genres.InventoryList {
		if co, err = importer.importInventoryList(external); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if co, err = importer.importLeafSku(external); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (importer importer) importInventoryList(
	listObject *sku.Transacted,
) (checkedOut *sku.CheckedOut, err error) {
	if err = genres.InventoryList.AssertGenre(listObject.GetGenre()); err != nil {
		err = errors.Wrap(err)
		return
	}

	blobSha := listObject.GetBlobSha()

	if !importer.envRepo.GetDefaultBlobStore().HasBlob(blobSha) {
		err = env_dir.ErrBlobMissing{
			DigestGetter: blobSha,
		}

		return
	}

	iter := importer.typedInventoryListBlobStore.StreamInventoryListBlobSkus(
		listObject,
	)

	for sk, errIter := range iter {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return
		}

		if _, err = importer.Import(
			sk,
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

	if checkedOut, err = importer.importLeafSku(
		listObject,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (importer importer) importLeafSku(
	external *sku.Transacted,
) (co *sku.CheckedOut, err error) {
	if importer.excludeObjects {
		// TODO remove this, it's expensive
		err = errors.ErrorWithStackf("skipping because objects are excluded")
		return
	}

	// TODO address this terrible hack? How should config objects be handled by
	// remotes?
	if external.GetGenre() == genres.Config {
		err = genres.MakeErrUnsupportedGenre(external.GetGenre())
		return
	}

	co = sku.GetCheckedOutPool().Get()

	sku.Resetter.ResetWith(co.GetSkuExternal(), external)

	// TODO set this as an importer option
	if co.GetSkuExternal().Metadata.RepoSig.IsEmpty() {
		if err = co.GetSkuExternal().Sign(
			importer.envRepo.GetConfigPrivate().Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = co.GetSkuExternal().CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if importer.indexObject != nil {
		_, err = importer.indexObject.ReadOneObjectIdTai(
			co.GetSkuExternal().GetObjectId(),
			co.GetSkuExternal().GetTai(),
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
		co.GetSkuExternal().GetObjectId(),
		co.GetSku(),
	); err != nil {
		if collections.IsErrNotFound(err) {
			if err = importer.storeObject.Commit(
				co.GetSkuExternal(),
				sku.CommitOptions{
					Clock:              co.GetSkuExternal(),
					StoreOptions:       importer.storeOptions,
					DontAddMissingTags: true,
					DontAddMissingType: true,
				},
			); err != nil {
				err = errors.WrapExcept(err, collections.ErrExists)
			}
		} else {
			err = errors.Wrapf(err, "ObjectId: %s", external.GetObjectId())
		}

		return
	}

	var commitOptions sku.CommitOptions

	// TODO extra commit option setting into its own function
	if importer.storeExternal != nil {
		if commitOptions, err = importer.storeExternal.MergeCheckedOut(
			co,
			importer.parentNegotiator,
			importer.allowMergeConflicts,
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		if co.GetState() == checked_out_state.Conflicted {
			if !importer.allowMergeConflicts {
				if err = importer.checkedOutPrinter(co); err != nil {
					err = errors.Wrap(err)
					return
				}

				return
			}
		}
	}

	commitOptions.Validate = false

	if err = importer.storeObject.Commit(
		co.GetSkuExternal(),
		commitOptions,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = importer.checkedOutPrinter(co); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (importer importer) ImportBlobIfNecessary(
	sk *sku.Transacted,
) (err error) {
	blobSha := sk.GetBlobSha()

	if importer.remoteBlobStore == nil {
		// when this is a dumb HTTP remote, we expect local to push the missing
		// objects to us after the import call

		n := int64(-1)

		if importer.envRepo.GetDefaultBlobStore().HasBlob(blobSha) {
			n = -2
		}

		if importer.blobCopierDelegate != nil {
			if err = importer.blobCopierDelegate(
				sku.BlobCopyResult{
					Transacted: sk,
					Digest:     blobSha,
					N:          n,
				},
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

		return
	}

	if !importer.blobGenres.Contains(sk.GetGenre()) {
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
				blobSha,
				&progressWriter,
			); err != nil {
				if errors.Is(err, &env_dir.ErrAlreadyExists{}) {
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
						Transacted: sk,
						Digest:     blobSha,
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
				blobSha,
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
