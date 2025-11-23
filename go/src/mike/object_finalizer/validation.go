package object_finalizer

import (
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

func (finalizer *Finalizer) ValidateIfNecessary(
	daughter *sku.Transacted,
	mother *sku.Transacted,
	options sku.CommitOptions,
	// typedBlobStores typed_blob_store.Stores,
) (err error) {
	if !options.Validate {
		return err
	}

	switch daughter.GetSku().GetGenre() {
	case genres.Type:
		// var repool interfaces.FuncRepool

		// if _, repool, _, err = typedBlobStores.Type.ParseTypedBlob(
		// 	daughter.GetType(),
		// 	daughter.GetSku().GetBlobDigest(),
		// ); err != nil {
		// 	err = errors.Wrap(err)
		// 	return err
		// }

		// defer repool()
	}

	return err
}
