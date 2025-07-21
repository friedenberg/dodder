package blob_stores

import (
	"bytes"
	"fmt"
	"io"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type localHashBucketed struct {
	buckets  []int
	config   blob_store_configs.ConfigLocalHashBucketed
	basePath string
	tempFS   env_dir.TemporaryFS
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

func (blobStore localHashBucketed) GetLocalBlobStore() interfaces.LocalBlobStore {
	return blobStore
}

func (blobStore localHashBucketed) makeEnvDirConfig() env_dir.Config {
	return env_dir.MakeConfig(
		env_dir.MakeHashBucketPathJoinFunc(blobStore.buckets),
		blobStore.config.GetBlobCompression(),
		blobStore.config.GetBlobEncryption(),
	)
}

func (blobStore localHashBucketed) HasBlob(
	sh interfaces.Digest,
) (ok bool) {
	if sh.GetDigest().IsNull() {
		ok = true
		return
	}

	path := env_dir.MakeHashBucketPathFromSha(
		sh,
		blobStore.buckets,
		blobStore.basePath,
	)
	ok = files.Exists(path)

	return
}

// TODO add support for other bucket sizes
func (blobStore localHashBucketed) AllBlobs() interfaces.SeqError[interfaces.Digest] {
	return func(yield func(interfaces.Digest, error) bool) {
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

func (blobStore localHashBucketed) BlobWriter() (w interfaces.ShaWriteCloser, err error) {
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
	sh interfaces.Digest,
) (readeCloser interfaces.ShaReadCloser, err error) {
	if sh.GetDigest().IsNull() {
		readeCloser = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	if readeCloser, err = blobStore.blobReaderFrom(
		sh,
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
	sh sha.ShaLike,
	path string,
) (readCloser sha.ReadCloser, err error) {
	if sh.GetDigest().IsNull() {
		readCloser = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	path = env_dir.MakeHashBucketPathFromSha(
		sh.GetDigest(),
		blobStore.buckets,
		path,
	)

	if readCloser, err = env_dir.NewFileReader(
		blobStore.makeEnvDirConfig(),
		path,
	); err != nil {
		if errors.IsNotExist(err) {
			shCopy := sha.GetPool().Get()
			shCopy.ResetWithShaLike(sh.GetDigest())

			err = env_dir.ErrBlobMissing{
				DigestGetter: shCopy,
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
