package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/echo/repo_blobs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type RepoStore struct {
	envRepo env_repo.Env
}

func MakeRepoStore(
	dirLayout env_repo.Env,
) RepoStore {
	return RepoStore{
		envRepo: dirLayout,
	}
}

func (store RepoStore) ReadTypedBlob(
	tipe ids.Type,
	blobSha interfaces.BlobId,
) (common repo_blobs.Blob, n int64, err error) {
	var reader interfaces.ReadCloseBlobIdGetter

	if reader, err = store.envRepo.GetDefaultBlobStore().BlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, reader)

	typedBlob := repo_blobs.TypedBlob{
		Type: tipe,
	}

	bufferedReader, repoolBufferedReader := pool.GetBufferedReader(reader)
	defer repoolBufferedReader()

	if n, err = repo_blobs.Coder.DecodeFrom(
		&typedBlob,
		bufferedReader,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	common = *typedBlob.Blob

	return
}

func (store RepoStore) WriteTypedBlob(
	tipe ids.Type,
	blob repo_blobs.Blob,
) (sh interfaces.BlobId, n int64, err error) {
	var writer interfaces.WriteCloseBlobIdGetter

	if writer, err = store.envRepo.GetDefaultBlobStore().BlobWriter(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, writer)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(writer)
	defer repoolBufferedWriter()

	if n, err = repo_blobs.Coder.EncodeTo(
		&repo_blobs.TypedBlob{
			Type: tipe,
			Blob: &blob,
		},
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh = writer.GetBlobId()

	return
}
