package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (store *Store) ToSliceFilesZettelen(
	cos sku.SkuTypeSet,
) (out []string, err error) {
	return quiter.DerivedValues(
		cos,
		func(col sku.SkuType) (e string, err error) {
			var fds *sku.FSItem

			if fds, err = store.ReadFSItemFromExternal(col.GetSkuExternal()); err != nil {
				err = errors.Wrap(err)
				return
			}

			e = fds.Object.GetPath()

			if e == "" {
				err = errors.MakeErrStopIteration()
				return
			}

			return
		},
	)
}

func (store *Store) ToSliceFilesBlobs(
	cos sku.SkuTypeSet,
) (out []string, err error) {
	return quiter.DerivedValues(
		cos,
		func(col sku.SkuType) (e string, err error) {
			var fds *sku.FSItem

			if fds, err = store.ReadFSItemFromExternal(col.GetSkuExternal()); err != nil {
				err = errors.Wrap(err)
				return
			}

			e = fds.Blob.GetPath()

			if e == "" {
				err = errors.MakeErrStopIteration()
				return
			}

			return
		},
	)
}
