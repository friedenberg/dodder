package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

var _ interfaces.NamedBlobAccess = Env{}

func MakeNamedBlobReaderOrNullReader(
	blobAccess interfaces.NamedBlobAccess,
	path string,
) (blobReader interfaces.BlobReader, err error) {
	if blobReader, err = blobAccess.MakeNamedBlobReader(path); err != nil {
		if errors.IsNotExist(err) {
			return env_dir.NewNopReader()
		} else {
			err = errors.Wrap(err)
			return blobReader, err
		}
	}

	return blobReader, err
}

func (env Env) MakeNamedBlobReader(path string) (interfaces.BlobReader, error) {
	return env_dir.NewFileReaderOrErrNotExist(env_dir.DefaultConfig, path)
}

func (env Env) MakeNamedBlobWriter(
	path string,
) (interfaces.BlobWriter, error) {
	return env_dir.NewMover(
		env_dir.DefaultConfig,
		env_dir.MoveOptions{
			FinalPathOrDir: path,
			TemporaryFS:    env.GetTempLocal(),
		},
	)
}
