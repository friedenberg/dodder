package inventory_list_store

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
)

type blobStoreV0 struct {
	envRepo        env_repo.Env
	lock           sync.Mutex
	blobType       ids.Type
	typedBlobStore inventory_list_coders.Closet

	interfaces.BlobStore
}

func (blobStore *blobStoreV0) getType() ids.Type {
	return blobStore.blobType
}

func (blobStore *blobStoreV0) getTypedBlobStore() inventory_list_coders.Closet {
	return blobStore.typedBlobStore
}

// TODO rename to ReadOneDigest
func (blobStore *blobStoreV0) ReadOneSha(
	id interfaces.BlobId,
) (object *sku.Transacted, err error) {
	var sh sha.Sha

	if err = sh.Set(digests.Format(id)); err != nil {
		err = errors.Wrap(err)
		return
	}

	var readCloser interfaces.ReadCloseBlobIdGetter

	if readCloser, err = blobStore.BlobStore.BlobReader(&sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(readCloser)
	defer repoolBufferedReader()

	if object, err = blobStore.typedBlobStore.ReadInventoryListObject(
		blobStore.envRepo,
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

	var blobStoreWriteCloser interfaces.WriteCloseBlobIdGetter

	if blobStoreWriteCloser, err = blobStore.BlobStore.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = blobStore.blobType

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
		blobStoreWriteCloser,
	)
	defer repoolBufferedWriter()

	if err = object.CalculateObjectDigests(); err != nil {
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
		for sh, err := range blobStore.BlobStore.AllBlobs() {
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
