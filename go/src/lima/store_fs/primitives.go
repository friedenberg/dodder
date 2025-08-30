package store_fs

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// Internal may be nil, which means that the external is hydrated without an
// overlay.
func (store *Store) HydrateExternalFromItem(
	options sku.CommitOptions,
	item *sku.FSItem,
	internal *sku.Transacted,
	external *sku.Transacted,
) (err error) {
	if internal != nil {
		external.ObjectId.ResetWith(&internal.ObjectId)
	}

	if err = item.WriteToSku(
		external,
		store.envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var m checkout_mode.Mode

	if m, err = item.GetCheckoutModeOrError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	switch m {
	case checkout_mode.BlobOnly:
		if err = store.readOneExternalBlob(external, internal, item); err != nil {
			err = errors.Wrap(err)
			return
		}

	case checkout_mode.MetadataOnly, checkout_mode.MetadataAndBlob:
		if item.Object.IsStdin() {
			if err = store.ReadOneExternalObjectReader(os.Stdin, external); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			if err = store.readOneExternalObject(external, internal, item); err != nil {
				err = errors.Wrap(err)
				return
			}
		}

	case checkout_mode.BlobRecognized:
		object_metadata.Resetter.ResetWith(
			external.GetMetadata(),
			internal.GetMetadata(),
		)

	default:
		err = checkout_mode.MakeErrInvalidCheckoutModeMode(m)
		return
	}

	if options.Clock == nil {
		options.Clock = item
	}

	if err = store.WriteFSItemToExternal(item, external); err != nil {
		err = errors.Wrap(err)
		return
	}

	// Don't apply the proto object as that would artificially create deltas
	options.StoreOptions.ApplyProto = false

	if err = store.storeSupplies.Commit(
		external,
		options,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// Internal can be nil which means that no overlaying is done.
func (store *Store) readOneExternalObject(
	external *sku.Transacted,
	internal *sku.Transacted,
	item *sku.FSItem,
) (err error) {
	if internal != nil {
		object_metadata.Resetter.ResetWith(
			external.GetMetadata(),
			internal.GetMetadata(),
		)
	}

	var f *os.File

	if f, err = files.Open(item.Object.GetPath()); err != nil {
		err = errors.Wrapf(err, "Item: %s", item.Debug())
		return
	}

	defer errors.DeferredCloser(&err, f)

	if err = store.ReadOneExternalObjectReader(f, external); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *Store) readOneExternalBlob(
	external *sku.Transacted,
	internal *sku.Transacted,
	item *sku.FSItem,
) (err error) {
	object_metadata.Resetter.ResetWith(
		&external.Metadata,
		internal.GetMetadata(),
	)

	// TODO use cache
	{
		var writeCloser interfaces.WriteCloseMarklIdGetter

		if writeCloser, err = store.envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, writeCloser)

		var file *os.File

		if file, err = files.OpenExclusiveReadOnly(
			item.Blob.GetPath(),
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, file)

		if _, err = io.Copy(writeCloser, file); err != nil {
			err = errors.Wrap(err)
			return
		}

		markl.SetDigester(
			external.GetMetadata().GetBlobDigestMutable(),
			writeCloser,
		)
	}

	return
}
