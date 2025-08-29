package store_fs

import (
	"os"
	"os/exec"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO combine with other method in this file
// Makes hard assumptions about the availability of the blobs associated with
// the *sku.CheckedOut.
func (store *Store) MergeCheckedOut(
	co *sku.CheckedOut,
	parentNegotiator sku.ParentNegotiator,
	allowMergeConflicts bool,
) (commitOptions sku.CommitOptions, err error) {
	commitOptions.StoreOptions = sku.GetStoreOptionsImport()

	// TODO determine why the internal can ever be null
	if co.GetSku().Metadata.GetObjectDigest().IsNull() || allowMergeConflicts {
		return
	}

	var conflicts checkout_mode.Mode

	// TODO add checkout_mode.BlobOnly
	if merkle.Equals(
		co.GetSku().Metadata.GetObjectDigest(),
		co.GetSkuExternal().Metadata.GetObjectDigest(),
	) {
		commitOptions.StoreOptions = sku.StoreOptions{}
		return
	} else if co.GetSku().Metadata.EqualsSansTai(&co.GetSkuExternal().Metadata) {
		if !co.GetSku().Metadata.Tai.Less(co.GetSkuExternal().Metadata.Tai) {
			// TODO implement retroactive change
		}

		return
	} else if merkle.Equals(co.GetSku().Metadata.GetBlobDigest(), co.GetSkuExternal().Metadata.GetBlobDigest()) {
		conflicts = checkout_mode.MetadataOnly
	} else {
		conflicts = checkout_mode.MetadataAndBlob
	}

	// TODO write conflicts
	switch conflicts {
	case checkout_mode.BlobOnly:
	case checkout_mode.MetadataOnly:
	case checkout_mode.MetadataAndBlob:
	default:
	}

	conflicted := sku.Conflicted{
		CheckedOut: co,
		Local:      co.GetSku(),
		Remote:     co.GetSkuExternal(),
	}

	if err = conflicted.FindBestCommonAncestor(parentNegotiator); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = conflicted.Remote.SetMother(conflicted.Base); err != nil {
		err = errors.Wrap(err)
		return
	}

	var skuReplacement *sku.Transacted

	// TODO pass mode / conflicts
	if skuReplacement, err = store.MakeMergedTransacted(
		conflicted,
	); err != nil {
		if sku.IsErrMergeConflict(err) {
			err = nil

			if !allowMergeConflicts {
				if err = store.GenerateConflictMarker(
					conflicted,
					conflicted.CheckedOut,
				); err != nil {
					err = errors.Wrap(err)
					return
				}
			}

			co.SetState(checked_out_state.Conflicted)
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	sku.TransactedResetter.ResetWith(co.GetSkuExternal(), skuReplacement)

	return
}

func (store *Store) Merge(conflicted sku.Conflicted) (err error) {
	var original *sku.FSItem

	if original, err = store.ReadFSItemFromExternal(
		conflicted.CheckedOut.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var skuReplacement *sku.Transacted

	if skuReplacement, err = store.MakeMergedTransacted(conflicted); err != nil {
		if sku.IsErrMergeConflict(err) {
			if err = store.GenerateConflictMarker(
				conflicted,
				conflicted.CheckedOut,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	if original.FDs.Len() == 0 {
		// generate check out item
		// TODO if original is empty, it means this was not a checked out
		// conflict but
		// a remote conflict
	}

	var replacement *sku.FSItem

	if replacement, err = store.ReadFSItemFromExternal(skuReplacement); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !original.Object.IsEmpty() && !replacement.Object.IsEmpty() {
		if err = files.Rename(
			replacement.Object.GetPath(),
			original.Object.GetPath(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if !original.Blob.IsEmpty() && !replacement.Blob.IsEmpty() {
		if err = files.Rename(
			replacement.Blob.GetPath(),
			original.Blob.GetPath(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (store *Store) checkoutConflictedForMerge(
	tm sku.Conflicted,
	mode checkout_mode.Mode,
) (local, base, remote *sku.FSItem, err error) {
	if _, local, err = store.checkoutOneForMerge(mode, tm.Local); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, base, err = store.checkoutOneForMerge(mode, tm.Base); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, remote, err = store.checkoutOneForMerge(mode, tm.Remote); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) MakeMergedTransacted(
	conflicted sku.Conflicted,
) (merged *sku.Transacted, err error) {
	// tags have to be manually merged at the moment, even though there is a
	// simple algorithm to merge them automatically. this is because the tags
	// get merged before running diff3, but then local, base, and remote have
	// the merged tag set, and then their signatures become incorrect.
	// if err = conflicted.MergeTags(); err != nil {
	// 	err = errors.Wrap(err)
	// 	return
	// }

	var localItem, baseItem, remoteItem *sku.FSItem

	inlineBlob := conflicted.IsAllInlineType(store.config)

	mode := checkout_mode.MetadataAndBlob

	if !inlineBlob {
		mode = checkout_mode.MetadataOnly
	}

	if localItem, baseItem, remoteItem, err = store.checkoutConflictedForMerge(
		conflicted,
		mode,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var mergedItem *sku.FSItem
	var diff3Error error

	mergedItem, diff3Error = store.runDiff3(
		localItem,
		baseItem,
		remoteItem,
	)

	if diff3Error != nil {
		err = errors.Wrap(diff3Error)
		return
	}

	localItem.ResetWith(mergedItem)

	merged = GetExternalPool().Get()

	merged.ObjectId.ResetWith(&conflicted.GetSku().ObjectId)

	if err = store.WriteFSItemToExternal(localItem, merged); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.HydrateExternalFromItem(
		sku.CommitOptions{
			StoreOptions: sku.StoreOptions{
				UpdateTai: true,
			},
		},
		mergedItem,
		conflicted.GetSku(),
		conflicted.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) checkoutOneForMerge(
	mode checkout_mode.Mode,
	sk *sku.Transacted,
) (co *sku.CheckedOut, i *sku.FSItem, err error) {
	if sk == nil {
		i = &sku.FSItem{}
		i.Reset()
		return
	}

	options := checkout_options.Options{
		CheckoutMode: mode,
		OptionsWithoutMode: checkout_options.OptionsWithoutMode{
			Force: true,
			StoreSpecificOptions: CheckoutOptions{
				AllowConflicted: true,
				Path:            PathOptionTempLocal,
				// TODO handle binary blobs
				ForceInlineBlob: true,
			},
		},
	}

	co = GetCheckedOutPool().Get()
	sku.Resetter.ResetWith(co.GetSku(), sk)

	if i, err = store.ReadFSItemFromExternal(co.GetSku()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.checkoutOneForReal(
		options,
		co,
		i,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.WriteFSItemToExternal(i, co.GetSkuExternal()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) GenerateConflictMarker(
	conflicted sku.Conflicted,
	checkedOut *sku.CheckedOut,
) (err error) {
	var file *os.File

	if file, err = store.envRepo.GetTempLocal().FileTemp(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(file)
	defer repoolBufferedWriter()
	defer errors.DeferredFlusher(&err, bufferedWriter)

	blobStore := store.storeSupplies.BlobStore.InventoryList
	// TODO assert that left and right both have a mother sig

	for object := range conflicted.All() {
		if err = object.FinalizeAndSignIfNecessary(
			store.envRepo.GetConfigPrivate().Blob,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if _, err = blobStore.WriteBlobToWriter(
		store.envRepo,
		ids.DefaultOrPanic(genres.InventoryList),
		quiter.MakeSeqErrorFromSeq(conflicted.All()),
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var item *sku.FSItem

	if item, err = store.ReadFSItemFromExternal(
		checkedOut.GetSkuExternal(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = item.GenerateConflictFD(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if checkedOut.GetSkuExternal().GetGenre() == genres.Zettel {
		var zettelId ids.ZettelId

		if err = zettelId.Set(checkedOut.GetSkuExternal().GetObjectId().String()); err != nil {
			err = errors.Wrap(err)
			return
		}

		if _, err = env_dir.MakeDirIfNecessaryForStringerWithHeadAndTail(
			zettelId,
			store.envRepo.GetCwd(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = os.Rename(
		file.Name(),
		item.Conflict.GetPath(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	checkedOut.SetState(checked_out_state.Conflicted)

	return
}

func (store *Store) RunMergeTool(
	tool []string,
	conflicted sku.Conflicted,
) (checkedOut *sku.CheckedOut, err error) {
	if len(tool) == 0 {
		err = errors.ErrorWithStackf("no utility provided")
		return
	}

	checkedOut = conflicted.CheckedOut

	inlineBlob := conflicted.IsAllInlineType(store.config)

	mode := checkout_mode.MetadataAndBlob

	if !inlineBlob {
		mode = checkout_mode.MetadataOnly
	}

	var localItem, baseItem, remoteItem *sku.FSItem

	if localItem, baseItem, remoteItem, err = store.checkoutConflictedForMerge(
		conflicted,
		mode,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var skuReplacement *sku.Transacted
	var replacement *sku.FSItem

	if skuReplacement, err = store.MakeMergedTransacted(conflicted); err != nil {
		var mergeConflict *sku.ErrMergeConflict

		if errors.As(err, &mergeConflict) {
			err = nil
			replacement = &mergeConflict.FSItem
		} else {
			err = errors.Wrap(err)
			return
		}
	} else {
		if replacement, err = store.ReadFSItemFromExternal(skuReplacement); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	tool = append(
		tool,
		localItem.Object.GetPath(),
		baseItem.Object.GetPath(),
		remoteItem.Object.GetPath(),
		replacement.Object.GetPath(),
	)

	// TODO merge blobs

	cmd := exec.Command(tool[0], tool[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	ui.Log().Print(cmd.Env)

	if err = cmd.Run(); err != nil {
		err = errors.Wrapf(err, "Cmd: %q", tool)
		return
	}

	external := GetExternalPool().Get()
	defer GetExternalPool().Put(external)

	external.ObjectId.ResetWith(&checkedOut.GetSkuExternal().ObjectId)

	if err = store.WriteFSItemToExternal(localItem, external); err != nil {
		err = errors.Wrap(err)
		return
	}

	var f *os.File

	if f, err = files.Open(replacement.Object.GetPath()); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO open blob

	defer errors.DeferredCloser(&err, f)

	if err = store.ReadOneExternalObjectReader(f, external); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = store.DeleteCheckedOut(
		conflicted.CheckedOut,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	checkedOut = GetCheckedOutPool().Get()

	sku.TransactedResetter.ResetWith(checkedOut.GetSkuExternal(), external)

	return
}
