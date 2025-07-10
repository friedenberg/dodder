package inventory_list_store

import (
	"io"
	"iter"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

// TODO add triple_hyphen_io2 coder
type objectBlobStoreV1 struct {
	envRepo        env_repo.Env
	pathLog        string
	blobType       ids.Type
	typedBlobStore typed_blob_store.InventoryList

	interfaces.LocalBlobStore
}

func (store *objectBlobStoreV1) getType() ids.Type {
	return store.blobType
}

func (store *objectBlobStoreV1) getTypedBlobStore() typed_blob_store.InventoryList {
	return store.typedBlobStore
}

func (store *objectBlobStoreV1) ReadOneSha(
	id interfaces.Stringer,
) (object *sku.Transacted, err error) {
	var sh sha.Sha

	if err = sh.Set(id.String()); err != nil {
		err = errors.Wrap(err)
		return
	}

	var readCloser sha.ReadCloser

	if readCloser, err = store.LocalBlobStore.BlobReader(&sh); err != nil {
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

func (store *objectBlobStoreV1) WriteInventoryListObject(
	object *sku.Transacted,
) (err error) {
	var blobStoreWriteCloser interfaces.ShaWriteCloser

	if blobStoreWriteCloser, err = store.LocalBlobStore.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = store.blobType

	var file *os.File

	if file, err = files.OpenExclusiveWriteOnlyAppend(store.pathLog); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)
	defer errors.Deferred(&err, file.Sync)

	bufferedWriter := ohio.BufferedWriter(
		io.MultiWriter(blobStoreWriteCloser, file),
	)
	defer pool.GetBufioWriter().Put(bufferedWriter)

	if err = object.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = object.Sign(
		store.envRepo.GetConfigPrivate().Blob,
	); err != nil {
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

	// TODO why do we CalculateObjectShas twice?
	if err = object.CalculateObjectShas(); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Printf(
		"saved inventory list: %q",
		sku.String(object),
	)

	return
}

func (store *objectBlobStoreV1) IterAllInventoryLists() iter.Seq2[*sku.Transacted, error] {
	return func(yield func(*sku.Transacted, error) bool) {
		var file *os.File

		{
			var err error

			if file, err = files.OpenReadOnly(store.pathLog); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		seq := store.typedBlobStore.AllDecodedObjectsFromStream(
			file,
		)

		for sk, err := range seq {
			if err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}
			}

			if !yield(sk, nil) {
				return
			}
		}

		if err := file.Close(); err != nil {
			yield(nil, errors.Wrap(err))
			return
		}
	}
}
