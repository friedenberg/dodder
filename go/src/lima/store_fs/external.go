package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) GetCheckoutMode(
	el sku.ExternalLike,
) (m checkout_mode.Mode, err error) {
	var fds *sku.FSItem

	if fds, err = store.ReadFSItemFromExternal(el); err != nil {
		err = errors.Wrap(err)
		return m, err
	}

	m = fds.GetCheckoutMode()

	return m, err
}

func (store *Store) GetCheckoutModeOrError(
	el sku.ExternalLike,
) (m checkout_mode.Mode, err error) {
	var fds *sku.FSItem

	if fds, err = store.ReadFSItemFromExternal(el); err != nil {
		err = errors.Wrap(err)
		return m, err
	}

	if m, err = fds.GetCheckoutModeOrError(); err != nil {
		err = errors.Wrap(err)
		return m, err
	}

	return m, err
}

func (store *Store) GetConflictOrError(
	el sku.ExternalLike,
) (f *fd.FD, err error) {
	var fds *sku.FSItem

	if fds, err = store.ReadFSItemFromExternal(el); err != nil {
		err = errors.Wrap(err)
		return f, err
	}

	f = &fds.Conflict

	return f, err
}

func (store *Store) GetObjectOrError(
	el sku.ExternalLike,
) (f *fd.FD, err error) {
	var item *sku.FSItem

	if item, err = store.ReadFSItemFromExternal(el); err != nil {
		err = errors.Wrap(err)
		return f, err
	}

	f = &item.Object

	return f, err
}

func (store *Store) UpdateTransactedFromBlobs(
	el sku.ExternalLike,
) (err error) {
	sk := el.GetSku()

	var item *sku.FSItem

	if item, err = store.ReadFSItemFromExternal(sk); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if sk.Metadata.Description.IsEmpty() {
		sorted := quiter.ElementsSorted(
			item.FDs,
			func(a, b *fd.FD) bool {
				return a.GetPath() < b.GetPath()
			},
		)

		for _, f := range sorted {
			desc := f.FileNameSansExt()

			if err = sk.Metadata.Description.Set(desc); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}
	}

	if !item.Blob.IsEmpty() {
		blobFD := &item.Blob
		ext := blobFD.ExtSansDot()
		typFromExtension := store.config.GetTypeStringFromExtension(ext)

		if typFromExtension == "" {
			typFromExtension = ext
		}

		if typFromExtension != "" {
			if err = sk.Metadata.Type.Set(typFromExtension); err != nil {
				err = errors.Wrapf(err, "Path: %s", blobFD.GetPath())
				return err
			}
		}
	}

	if err = store.WriteFSItemToExternal(item, sk); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
