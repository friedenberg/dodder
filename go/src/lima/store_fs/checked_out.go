package store_fs

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (s *Store) ToSliceFilesZettelen(
	cos sku.SkuTypeSet,
) (out []string, err error) {
	return quiter.DerivedValues(
		cos,
		func(col sku.SkuType) (e string, err error) {
			var fds *sku.FSItem

			if fds, err = s.ReadFSItemFromExternal(col.GetSkuExternal()); err != nil {
				err = errors.Wrap(err)
				return
			}

			e = fds.Object.GetPath()

			if e == "" {
				err = collections.MakeErrStopIteration()
				return
			}

			return
		},
	)
}

func (s *Store) ToSliceFilesBlobs(
	cos sku.SkuTypeSet,
) (out []string, err error) {
	return quiter.DerivedValues(
		cos,
		func(col sku.SkuType) (e string, err error) {
			var fds *sku.FSItem

			if fds, err = s.ReadFSItemFromExternal(col.GetSkuExternal()); err != nil {
				err = errors.Wrap(err)
				return
			}

			e = fds.Blob.GetPath()

			if e == "" {
				err = collections.MakeErrStopIteration()
				return
			}

			return
		},
	)
}
