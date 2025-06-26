package blob_store

import (
	"bytes"
	"io"
	"iter"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/id"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

type LocalBlobStore interface {
	interfaces.LocalBlobStore
	Mover() (*env_dir.Mover, error)
}

type storeShardedFiles struct {
	env_dir.Config
	basePath string
	tempFS   env_dir.TemporaryFS
}

func MakeShardedFilesStore(
	basePath string,
	config env_dir.Config,
	tempFS env_dir.TemporaryFS,
) storeShardedFiles {
	return storeShardedFiles{
		Config:   config,
		basePath: basePath,
		tempFS:   tempFS,
	}
}

func (s storeShardedFiles) GetBlobStore() interfaces.BlobStore {
	return s
}

func (s storeShardedFiles) GetLocalBlobStore() interfaces.LocalBlobStore {
	return s
}

func (s storeShardedFiles) HasBlob(
	sh interfaces.Sha,
) (ok bool) {
	if sh.GetShaLike().IsNull() {
		ok = true
		return
	}

	p := id.Path(sh.GetShaLike(), s.basePath)
	ok = files.Exists(p)

	return
}

func (s storeShardedFiles) AllBlobs() iter.Seq2[interfaces.Sha, error] {
	return func(yield func(interfaces.Sha, error) bool) {
		var sh sha.Sha

		for path, err := range files.DirNamesLevel2(s.basePath) {
			if err != nil {
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

			if !yield(&sh, nil) {
				return
			}
		}
	}
}

func (store storeShardedFiles) BlobWriter() (w interfaces.ShaWriteCloser, err error) {
	if w, err = store.blobWriterTo(store.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (store storeShardedFiles) Mover() (mover *env_dir.Mover, err error) {
	if mover, err = store.blobWriterTo(store.basePath); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s storeShardedFiles) BlobReader(
	sh interfaces.Sha,
) (r interfaces.ShaReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		r = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	if r, err = s.blobReaderFrom(sh, s.basePath); err != nil {
		if !env_dir.IsErrBlobMissing(err) {
			err = errors.Wrap(err)
		}

		return
	}

	return
}

func (store storeShardedFiles) blobWriterTo(
	path string,
) (mover *env_dir.Mover, err error) {
	options := env_dir.MoveOptions{
		Config:                   store.Config,
		FinalPath:                path,
		GenerateFinalPathFromSha: true,
		TemporaryFS:              store.tempFS,
	}

	if mover, err = env_dir.NewMover(options); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (s storeShardedFiles) blobReaderFrom(
	sh sha.ShaLike,
	p string,
) (r sha.ReadCloser, err error) {
	if sh.GetShaLike().IsNull() {
		r = sha.MakeNopReadCloser(io.NopCloser(bytes.NewReader(nil)))
		return
	}

	p = id.Path(sh.GetShaLike(), p)

	o := env_dir.FileReadOptions{
		Config: s.Config,
		Path:   p,
	}

	if r, err = env_dir.NewFileReader(o); err != nil {
		if errors.IsNotExist(err) {
			shCopy := sha.GetPool().Get()
			shCopy.ResetWithShaLike(sh.GetShaLike())

			err = env_dir.ErrBlobMissing{
				ShaGetter: shCopy,
				Path:      p,
			}
		} else {
			err = errors.Wrapf(
				err,
				"Path: %q, Compression: %q",
				p,
				s.GetBlobCompression(),
			)
		}

		return
	}

	return
}
