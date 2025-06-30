package inventory_list_store

import (
	"iter"
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

type objectBlobStoreV0 struct {
	lock           sync.Mutex
	blobType       ids.Type
	typedBlobStore typed_blob_store.InventoryList
	blobStore      interfaces.LocalBlobStore
}

func (store *objectBlobStoreV0) getType() ids.Type {
	return store.blobType
}

func (store *objectBlobStoreV0) getTypedBlobStore() typed_blob_store.InventoryList {
	return store.typedBlobStore
}

func (store *objectBlobStoreV0) ReadOneSha(
	id interfaces.Stringer,
) (object *sku.Transacted, err error) {
	var sh sha.Sha

	if err = sh.Set(id.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var readCloser sha.ReadCloser

	if readCloser, err = store.blobStore.BlobReader(&sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader := ohio.BufferedReader(readCloser)
	defer pool.GetBufioReader().Put(bufferedReader)

	if object, err = store.typedBlobStore.ReadInventoryListObject(
		store.blobType,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store *objectBlobStoreV0) WriteInventoryListObject(
	object *sku.Transacted,
) (err error) {
	store.lock.Lock()
	defer store.lock.Unlock()

	var blobStoreWriteCloser interfaces.ShaWriteCloser

	if blobStoreWriteCloser, err = store.blobStore.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = store.blobType

	bufferedWriter := ohio.BufferedWriter(blobStoreWriteCloser)
	defer pool.GetBufioWriter().Put(bufferedWriter)

	if err = object.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = store.typedBlobStore.WriteObjectToWriter(
		store.blobType,
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

func (objectBlobStore *objectBlobStoreV0) IterAllInventoryLists() iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		for sh, err := range objectBlobStore.blobStore.AllBlobs() {
			if err != nil {
				if !yield(nil, err) {
					return
				}
			}

			var decodedList *sku.Transacted

			if decodedList, err = objectBlobStore.ReadOneSha(sh); err != nil {
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
