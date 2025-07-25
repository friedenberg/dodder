package blob_stores

import (
	"bytes"
	"fmt"
	"io"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type localHashBucketed struct {
	// TODO move to config
	envDigest sha.Env
	buckets   []int
	config    blob_store_configs.ConfigLocalHashBucketed
	basePath  string
	tempFS    env_dir.TemporaryFS
}

func makeLocalHashBucketed(
	ctx interfaces.Context,
	basePath string,
	config blob_store_configs.ConfigLocalHashBucketed,
	tempFS env_dir.TemporaryFS,
) (localHashBucketed, error) {
	// TODO validate
	store := localHashBucketed{
		buckets:  config.GetHashBuckets(),
		config:   config,
		basePath: basePath,
		tempFS:   tempFS,
	}

	return store, nil
}

func (blobStore localHashBucketed) GetBlobStoreDescription() string {
	return fmt.Sprintf("TODO: local-git-like")
}

func (blobStore localHashBucketed) GetBlobIOWrapper() interfaces.BlobIOWrapper {
	return blobStore.config
}

func (blobStore localHashBucketed) GetLocalBlobStore() interfaces.BlobStore {
	return blobStore
}

func (blobStore localHashBucketed) makeEnvDirConfig() env_dir.Config {
	return env_dir.MakeConfig(
		// TODO move to config
		sha.Env{},
		env_dir.MakeHashBucketPathJoinFunc(blobStore.buckets),
		blobStore.config.GetBlobCompression(),
		blobStore.config.GetBlobEncryption(),
	)
}

func (blobStore localHashBucketed) HasBlob(
	digest interfaces.BlobId,
) (ok bool) {
	if digest.GetBlobId().IsNull() {
		ok = true
		return
	}

	path := env_dir.MakeHashBucketPathFromSha(
		digest,
		blobStore.buckets,
		blobStore.basePath,
	)
	ok = files.Exists(path)

	return
}

// TODO add support for other bucket sizes and digest types
func (blobStore localHashBucketed) AllBlobs() interfaces.SeqError[interfaces.BlobId] {
	return func(yield func(interfaces.BlobId, error) bool) {
		var sh sha.Sha

		for path, err := range files.DirNamesLevel2(blobStore.basePath) {
			if errors.IsErrno(err, syscall.ENOTDIR) {
				err = nil
				continue
			}

			if err != nil {
				err = errors.Wrap(err)
				if !yield(nil, err) {
					return
				}
			}

			if err = sh.SetFromPath(path); err != nil {
				err = errors.Wrap(err)
				if !yield(nil, err) {
					return
				}
			}

			if sh.IsNull() {
				continue
			}

			if !yield(&sh, nil) {
				return
			}
		}
	}
}

func (blobStore localHashBucketed) BlobWriter() (w interfaces.WriteCloseBlobIdGetter, err error) {
	if w, err = blobStore.blobWriterTo(blobStore.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore localHashBucketed) Mover() (mover interfaces.Mover, err error) {
	if mover, err = blobStore.blobWriterTo(blobStore.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore localHashBucketed) BlobReader(
	digest interfaces.BlobId,
) (readCloser interfaces.ReadCloseBlobIdGetter, err error) {
	if digest.GetBlobId().IsNull() {
		readCloser = digests.MakeNopReadCloser(
			blobStore.envDigest,
			io.NopCloser(bytes.NewReader(nil)),
		)
		return
	}

	if readCloser, err = blobStore.blobReaderFrom(
		digest,
		blobStore.basePath,
	); err != nil {
		if !env_dir.IsErrBlobMissing(err) {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (blobStore localHashBucketed) blobWriterTo(
	path string,
) (mover interfaces.Mover, err error) {
	if mover, err = env_dir.NewMover(
		blobStore.makeEnvDirConfig(),
		env_dir.MoveOptions{
			FinalPath:                path,
			GenerateFinalPathFromSha: true,
			TemporaryFS:              blobStore.tempFS,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore localHashBucketed) blobReaderFrom(
	digest interfaces.BlobId,
	path string,
) (readCloser interfaces.ReadCloseBlobIdGetter, err error) {
	if digest.IsNull() {
		readCloser = digests.MakeNopReadCloser(
			blobStore.envDigest,
			io.NopCloser(bytes.NewReader(nil)),
		)
		return
	}

	path = env_dir.MakeHashBucketPathFromSha(
		digest.GetBlobId(),
		blobStore.buckets,
		path,
	)

	if readCloser, err = env_dir.NewFileReader(
		blobStore.makeEnvDirConfig(),
		path,
	); err != nil {
		if errors.IsNotExist(err) {
			shCopy := sha.GetPool().Get()
			shCopy.ResetWithShaLike(digest.GetBlobId())

			err = env_dir.ErrBlobMissing{
				BlobIdGetter: shCopy,
				Path:         path,
			}
		} else {
			err = errors.Wrapf(
				err,
				"Path: %q, Compression: %q",
				path,
				blobStore.config.GetBlobCompression(),
			)
		}

		return
	}

	return
}
