package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type BlobStore[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] struct {
	envRepo env_repo.Env
	pool    interfaces.Pool[BLOB, BLOB_PTR]
	Format[BLOB, BLOB_PTR]
	resetFunc func(BLOB_PTR)
}

func MakeBlobStore[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
](
	repoLayout env_repo.Env,
	format Format[BLOB, BLOB_PTR],
	resetFunc func(BLOB_PTR),
) (blobStore *BlobStore[BLOB, BLOB_PTR]) {
	blobStore = &BlobStore[BLOB, BLOB_PTR]{
		envRepo:   repoLayout,
		pool:      pool.MakePool(nil, resetFunc),
		Format:    format,
		resetFunc: resetFunc,
	}

	return
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) GetBlob(
	blobId interfaces.BlobId,
) (blobPtr BLOB_PTR, repool interfaces.FuncRepool, err error) {
	var readCloser interfaces.ReadCloseBlobIdGetter

	if readCloser, err = blobStore.envRepo.GetDefaultBlobStore().BlobReader(
		blobId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	blobPtr = blobStore.pool.Get()

	if _, err = blobStore.DecodeFrom(blobPtr, readCloser); err != nil {
		err = errors.Wrapf(err, "BlobReader: %q", readCloser)
		return
	}

	actual := readCloser.GetBlobId()

	if !merkle_ids.Equals(actual, blobId) {
		err = errors.ErrorWithStackf(
			"expected sha %s but got %s",
			blobId,
			actual,
		)

		return
	}

	repool = func() {
		blobStore.pool.Put(blobPtr)
	}

	return
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) PutBlob(a BLOB_PTR) {
	// TODO-P2 implement pool
}

// TODO re-evaluate this strategy
func (blobStore *BlobStore[BLOB, BLOB_PTR]) SaveBlobText(
	o BLOB_PTR,
) (sh interfaces.BlobId, n int64, err error) {
	var writeCloser interfaces.WriteCloseBlobIdGetter

	if writeCloser, err = blobStore.envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if n, err = blobStore.EncodeTo(o, writeCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = writeCloser.GetBlobId()

	return
}
