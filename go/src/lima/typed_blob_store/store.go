package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type BlobStore[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] struct {
	envRepo env_repo.Env
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
		Format:    format,
		resetFunc: resetFunc,
	}

	return
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) GetBlob2(
	digest interfaces.BlobId,
) (blobPtr BLOB_PTR, repool interfaces.FuncRepool, err error) {
	var readCloser interfaces.ReadCloseDigester

	if readCloser, err = blobStore.envRepo.GetDefaultBlobStore().BlobReader(
		digest,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, readCloser)

	var blob BLOB
	blobPtr = BLOB_PTR(&blob)
	blobStore.resetFunc(blobPtr)

	if _, err = blobStore.DecodeFrom(blobPtr, readCloser); err != nil {
		err = errors.Wrapf(err, "BlobReader: %q", readCloser)
		return
	}

	actual := readCloser.GetBlobId()

	if !digests.Equals(actual, digest) {
		err = errors.ErrorWithStackf(
			"expected sha %s but got %s",
			digest,
			actual,
		)

		return
	}

	repool = func() {
		blobStore.PutBlob(blobPtr)
	}

	return
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) GetBlob(
	digest interfaces.BlobId,
) (a BLOB_PTR, err error) {
	a, _, err = blobStore.GetBlob2(digest)
	return
}

func (blobStore *BlobStore[BLOB, BLOB_PTR]) PutBlob(a BLOB_PTR) {
	// TODO-P2 implement pool
}

// TODO re-evaluate this strategy
func (blobStore *BlobStore[BLOB, BLOB_PTR]) SaveBlobText(
	o BLOB_PTR,
) (sh interfaces.BlobId, n int64, err error) {
	var writeCloser interfaces.WriteCloseDigester

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
