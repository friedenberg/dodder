package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/foxtrot/markl"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
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
		pool:      pool.Make(nil, resetFunc),
		Format:    format,
		resetFunc: resetFunc,
	}

	return blobStore
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) GetBlob(
	blobId interfaces.MarklId,
) (blobPtr BLOB_PTR, repool interfaces.FuncRepool, err error) {
	var readCloser interfaces.BlobReader

	if readCloser, err = blobStore.envRepo.GetDefaultBlobStore().MakeBlobReader(
		blobId,
	); err != nil {
		err = errors.Wrap(err)
		return blobPtr, repool, err
	}

	defer errors.DeferredCloser(&err, readCloser)

	blobPtr = blobStore.pool.Get()

	if _, err = blobStore.DecodeFrom(blobPtr, readCloser); err != nil {
		err = errors.Wrapf(err, "BlobReader: %q", readCloser)
		return blobPtr, repool, err
	}

	actual := readCloser.GetMarklId()

	if err = markl.AssertEqual(blobId, actual); err != nil {
		err = errors.Wrap(err)
		return blobPtr, repool, err
	}

	repool = func() {
		blobStore.pool.Put(blobPtr)
	}

	return blobPtr, repool, err
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) PutBlob(a BLOB_PTR) {
	// TODO-P2 implement pool
}

// TODO re-evaluate this strategy
func (blobStore *BlobStore[BLOB, BLOB_PTR]) SaveBlobText(
	o BLOB_PTR,
) (sh interfaces.MarklId, n int64, err error) {
	var writeCloser interfaces.BlobWriter

	if writeCloser, err = blobStore.envRepo.GetDefaultBlobStore().MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return sh, n, err
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if n, err = blobStore.EncodeTo(o, writeCloser); err != nil {
		err = errors.Wrap(err)
		return sh, n, err
	}

	sh = writeCloser.GetMarklId()

	return sh, n, err
}
