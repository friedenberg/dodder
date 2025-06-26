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
	dirLayout env_repo.Env
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
) (s *BlobStore[A, APtr]) {
	s = &BlobStore[A, APtr]{
		dirLayout: repoLayout,
		Format:    format,
		resetFunc: resetFunc,
	}

	return
}

func (s *BlobStore[A, APtr]) GetBlob(
	sh interfaces.Sha,
) (a APtr, err error) {
	var rc interfaces.ShaReadCloser

	if rc, err = s.dirLayout.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, rc)

	var a1 A
	a = APtr(&a1)
	s.resetFunc(a)

	if _, err = s.DecodeFrom(a, rc); err != nil {
		err = errors.Wrapf(err, "BlobReader: %q", rc)
		return
	}

	actual := rc.GetShaLike()

	if !actual.EqualsSha(sh) {
		err = errors.ErrorWithStackf("expected sha %s but got %s", sh, actual)
		return
	}

	return
}

func (s *BlobStore[A, APtr]) PutBlob(a APtr) {
	// TODO-P2 implement pool
}

func (h *BlobStore[A, APtr]) SaveBlobText(
	o APtr,
) (sh interfaces.Sha, n int64, err error) {
	var w sha.WriteCloser

	if w, err = h.dirLayout.BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, w)

	if n, err = h.EncodeTo(o, w); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = w.GetShaLike()

	return
}
