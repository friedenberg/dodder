package inventory_list_store

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type blobStoreV0 struct {
	lock           sync.Mutex
	blobType       ids.Type
	typedBlobStore typed_blob_store.InventoryList

	interfaces.LocalBlobStore
}

func (blobStore *blobStoreV0) getType() ids.Type {
	return blobStore.blobType
}

func (blobStore *blobStoreV0) getTypedBlobStore() typed_blob_store.InventoryList {
	return blobStore.typedBlobStore
}

func (blobStore *blobStoreV0) ReadOneSha(
	id interfaces.Stringer,
) (object *sku.Transacted, err error) {
	var sh sha.Sha

	if err = sh.Set(id.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var readCloser sha.ReadCloser

	if readCloser, err = blobStore.LocalBlobStore.BlobReader(&sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader := ohio.BufferedReader(readCloser)
	defer pool.GetBufioReader().Put(bufferedReader)

	if object, err = blobStore.typedBlobStore.ReadInventoryListObject(
		blobStore.blobType,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore *blobStoreV0) WriteInventoryListObject(
	object *sku.Transacted,
) (err error) {
	blobStore.lock.Lock()
	defer blobStore.lock.Unlock()

	var blobStoreWriteCloser interfaces.ShaWriteCloser

	if blobStoreWriteCloser, err = blobStore.LocalBlobStore.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = blobStore.blobType

	bufferedWriter := ohio.BufferedWriter(blobStoreWriteCloser)
	defer pool.GetBufioWriter().Put(bufferedWriter)

	if err = object.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = blobStore.typedBlobStore.WriteObjectToWriter(
		blobStore.blobType,
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Printf(
		"saved inventory list: %q",
		sku.String(object),
	)

	return
}

func (blobStore *blobStoreV0) IterAllInventoryLists() interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		for sh, err := range blobStore.LocalBlobStore.AllBlobs() {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			// TODO make changes to prevent null shas from ever being written
			if sh.IsNull() {
				continue
			}

			var decodedList *sku.Transacted

			if decodedList, err = blobStore.ReadOneSha(sh); err != nil {
				if !yield(nil, errors.Wrapf(err, "Sha: %q", sh)) {
					return
				}
			}

			if !yield(decodedList, nil) {
				return
			}
		}
	}
}
