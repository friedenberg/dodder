package blob_stores

import (
	"bytes"
	"io"
	"iter"
	"syscall"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/id"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type gitLikeBucketedConfig interface {
	blob_store_configs.Config
	GetLockInternalFiles() bool
}

type gitLikeBucketed struct {
	config   gitLikeBucketedConfig
	basePath string
	tempFS   env_dir.TemporaryFS
}

func makeGitLikeBucketedStore(
	basePath string,
	config gitLikeBucketedConfig,
	tempFS env_dir.TemporaryFS,
) gitLikeBucketed {
	return gitLikeBucketed{
		config:   config,
		basePath: basePath,
		tempFS:   tempFS,
	}
}

func (blobStore gitLikeBucketed) GetBlobStore() interfaces.BlobStore {
	return blobStore
}

func (blobStore gitLikeBucketed) GetLocalBlobStore() interfaces.LocalBlobStore {
	return blobStore
}

func (blobStore gitLikeBucketed) makeEnvDirConfig() env_dir.Config {
	return env_dir.Config{
		Compression:       blobStore.config.GetBlobCompression(),
		Encryption:        blobStore.config.GetBlobEncryption(),
		LockInternalFiles: blobStore.config.GetLockInternalFiles(),
	}
}

func (blobStore gitLikeBucketed) HasBlob(
	sh interfaces.Sha,
) (ok bool) {
	if sh.GetShaLike().IsNull() {
		ok = true
		return
	}

	p := id.Path(sh.GetShaLike(), blobStore.basePath)
	ok = files.Exists(p)

	return
}

func (blobStore gitLikeBucketed) AllBlobs() iter.Seq2[interfaces.Sha, error] {
	return func(yield func(interfaces.Sha, error) bool) {
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

func (blobStore gitLikeBucketed) BlobWriter() (w interfaces.ShaWriteCloser, err error) {
	if w, err = blobStore.blobWriterTo(blobStore.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore gitLikeBucketed) Mover() (mover interfaces.Mover, err error) {
	if mover, err = blobStore.blobWriterTo(blobStore.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore gitLikeBucketed) BlobReader(
	sh interfaces.Sha,
) (r interfaces.ShaReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		r = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	if r, err = blobStore.blobReaderFrom(sh, blobStore.basePath); err != nil {
		if !env_dir.IsErrBlobMissing(err) {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (blobStore gitLikeBucketed) blobWriterTo(
	path string,
) (mover *env_dir.Mover, err error) {
	options := env_dir.MoveOptions{
		Config:                   blobStore.makeEnvDirConfig(),
		FinalPath:                path,
		GenerateFinalPathFromSha: true,
		TemporaryFS:              blobStore.tempFS,
	}

	if mover, err = env_dir.NewMover(options); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobStore gitLikeBucketed) blobReaderFrom(
	sh sha.ShaLike,
	path string,
) (readCloser sha.ReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		readCloser = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	path = id.Path(sh.GetShaLike(), path)

	options := env_dir.FileReadOptions{
		Config: blobStore.makeEnvDirConfig(),
		Path:   path,
	}

	if readCloser, err = env_dir.NewFileReader(options); err != nil {
		if errors.IsNotExist(err) {
			shCopy := sha.GetPool().Get()
			shCopy.ResetWithShaLike(sh.GetShaLike())

			err = env_dir.ErrBlobMissing{
				ShaGetter: shCopy,
				Path:      path,
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
