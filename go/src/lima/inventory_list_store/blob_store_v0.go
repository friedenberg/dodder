package inventory_list_store

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_finalizer"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
)

type blobStoreV0 struct {
	envRepo                  env_repo.Env
	lock                     sync.Mutex
	blobType                 ids.Type
	listFormat               sku.ListCoder
	inventoryListCoderCloset inventory_list_coders.Closet
	finalizer                object_finalizer.Finalizer

	interfaces.BlobStore
}

func (blobStore *blobStoreV0) getType() ids.Type {
	return blobStore.blobType
}

func (blobStore *blobStoreV0) getFormat() sku.ListCoder {
	return blobStore.listFormat
}

func (blobStore *blobStoreV0) GetInventoryListCoderCloset() inventory_list_coders.Closet {
	return blobStore.inventoryListCoderCloset
}

func (blobStore *blobStoreV0) ReadOneBlobId(
	blobId interfaces.MarklId,
) (object *sku.Transacted, err error) {
	var readCloser interfaces.BlobReader

	if readCloser, err = blobStore.BlobStore.MakeBlobReader(blobId); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)
	defer repoolBufferedReader()

	if object, err = blobStore.inventoryListCoderCloset.ReadInventoryListObject(
		blobStore.envRepo,
		blobStore.blobType,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return object, err
	}

	return object, err
}

func (blobStore *blobStoreV0) WriteInventoryListObject(
	object *sku.Transacted,
) (err error) {
	blobStore.lock.Lock()
	defer blobStore.lock.Unlock()

	var blobStoreWriteCloser interfaces.BlobWriter

	if blobStoreWriteCloser, err = blobStore.BlobStore.MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = blobStore.blobType

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
		blobStoreWriteCloser,
	)
	defer repoolBufferedWriter()

	if err = blobStore.finalizer.CalculateObjectDigests(object); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if _, err = blobStore.inventoryListCoderCloset.WriteObjectToWriter(
		blobStore.blobType,
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	ui.Log().Printf(
		"saved inventory list: %q",
		sku.String(object),
	)

	return err
}

func (blobStore *blobStoreV0) AllInventoryLists() sku.Seq {
	return func(yield func(*sku.Transacted, error) bool) {
		for blobId, err := range blobStore.BlobStore.AllBlobs() {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			// TODO make changes to prevent null shas from ever being written
			if blobId.IsNull() {
				continue
			}

			var list *sku.Transacted

			if list, err = blobStore.ReadOneBlobId(blobId); err != nil {
				if !yield(nil, errors.Wrapf(err, "BlobId: %q", blobId)) {
					return
				}

				continue
			}

			if !yield(list, nil) {
				return
			}
		}
	}
}
