package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type BlobStore[
	A any,
	APtr interfaces.Ptr[A],
] struct {
	envRepo env_repo.Env
	Format[A, APtr]
	resetFunc func(APtr)
}

func MakeBlobStore[
	A any,
	APtr interfaces.Ptr[A],
](
	repoLayout env_repo.Env,
	format Format[A, APtr],
	resetFunc func(APtr),
) (blobStore *BlobStore[A, APtr]) {
	blobStore = &BlobStore[A, APtr]{
		envRepo:   repoLayout,
		Format:    format,
		resetFunc: resetFunc,
	}

	return
}

func (blobStore *BlobStore[A, APtr]) GetBlob(
	sh interfaces.Digest,
) (a APtr, err error) {
	var rc interfaces.ReadCloserDigester

	if rc, err = blobStore.envRepo.GetDefaultBlobStore().BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, rc)

	var a1 A
	a = APtr(&a1)
	blobStore.resetFunc(a)

	if _, err = blobStore.DecodeFrom(a, rc); err != nil {
		err = errors.Wrapf(err, "BlobReader: %q", rc)
		return
	}

	actual := rc.GetDigest()

	if !interfaces.DigestEquals(actual, sh) {
		err = errors.ErrorWithStackf("expected sha %s but got %s", sh, actual)
		return
	}

	return
}

func (blobStore *BlobStore[A, APtr]) PutBlob(a APtr) {
	// TODO-P2 implement pool
}

func (blobStore *BlobStore[A, APtr]) SaveBlobText(
	o APtr,
) (sh interfaces.Digest, n int64, err error) {
	var writeCloser sha.WriteCloser

	if writeCloser, err = blobStore.envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writeCloser)

	if n, err = blobStore.EncodeTo(o, writeCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = writeCloser.GetDigest()

	return
}
