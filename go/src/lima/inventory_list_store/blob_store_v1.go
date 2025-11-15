package inventory_list_store

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_finalizer"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
)

type blobStoreV1 struct {
	envRepo                  env_repo.Env
	pathLog                  string
	blobType                 ids.Type
	listFormat               sku.ListCoder
	inventoryListCoderCloset inventory_list_coders.Closet
	finalizer                object_finalizer.Finalizer

	interfaces.BlobStore
}

func (blobStore *blobStoreV1) GetObjectFinalizer() object_finalizer.Finalizer {
	return blobStore.finalizer
}

func (blobStore *blobStoreV1) getType() ids.Type {
	return blobStore.blobType
}

func (blobStore *blobStoreV1) getFormat() sku.ListCoder {
	return blobStore.listFormat
}

func (blobStore *blobStoreV1) GetInventoryListCoderCloset() inventory_list_coders.Closet {
	return blobStore.inventoryListCoderCloset
}

func (blobStore *blobStoreV1) ReadOneBlobId(
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

func (blobStore *blobStoreV1) WriteInventoryListObject(
	object *sku.Transacted,
) (err error) {
	var blobStoreWriteCloser interfaces.BlobWriter

	if blobStoreWriteCloser, err = blobStore.BlobStore.MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, blobStoreWriteCloser)

	object.Metadata.Type = blobStore.blobType

	var file *os.File

	if file, err = files.OpenExclusiveWriteOnlyAppend(blobStore.pathLog); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, file)
	defer errors.Deferred(&err, file.Sync)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
		io.MultiWriter(blobStoreWriteCloser, file),
	)
	defer repoolBufferedWriter()

	if err = blobStore.finalizer.FinalizeAndSignOverwrite(
		object,
		blobStore.envRepo.GetConfigPrivate().Blob,
	); err != nil {
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

func (blobStore *blobStoreV1) AllInventoryLists() sku.Seq {
	return func(yield func(*sku.Transacted, error) bool) {
		var file *os.File

		{
			var err error

			if file, err = files.OpenReadOnly(blobStore.pathLog); err != nil {
				yield(nil, errors.Wrap(err))
				return
			}
		}

		defer errors.ContextMustClose(blobStore.envRepo, file)

		seq := blobStore.inventoryListCoderCloset.AllDecodedObjectsFromStream(
			file,
			nil,
		)

		for object, err := range seq {
			if err != nil {
				if !yield(nil, errors.Wrap(err)) {
					return
				}
			}

			if !yield(object, nil) {
				return
			}
		}
	}
}
