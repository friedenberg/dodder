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
	envRepo env_repo.Env,
) RepoStore {
	return RepoStore{
		envRepo: envRepo,
	}
}

func (store RepoStore) ReadTypedBlob(
	tipe ids.Type,
	blobSha interfaces.MarklId,
) (common repo_blobs.Blob, n int64, err error) {
	var reader interfaces.BlobReader

	if reader, err = store.envRepo.GetDefaultBlobStore().MakeBlobReader(blobSha); err != nil {
		err = errors.Wrap(err)
		return common, n, err
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
		return common, n, err
	}

	common = typedBlob.Blob

	return common, n, err
}

func (store RepoStore) WriteTypedBlob(
	tipe ids.Type,
	blob repo_blobs.Blob,
) (sh interfaces.MarklId, n int64, err error) {
	var writer interfaces.BlobWriter

	if writer, err = store.envRepo.GetDefaultBlobStore().MakeBlobWriter(nil); err != nil {
		err = errors.Wrap(err)
		return sh, n, err
	}

	defer errors.DeferredCloser(&err, writer)

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(writer)
	defer repoolBufferedWriter()

	if n, err = repo_blobs.Coder.EncodeTo(
		&repo_blobs.TypedBlob{
			Type: tipe,
			Blob: blob,
		},
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return sh, n, err
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return sh, n, err
	}

	sh = writer.GetMarklId()

	return sh, n, err
}
