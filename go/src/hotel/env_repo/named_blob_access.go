package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

var _ interfaces.NamedBlobAccess = Env{}

func (env Env) MakeNamedBlobReader(path string) (interfaces.BlobReader, error) {
	return env_dir.NewFileReaderOrEmptyBytesReader(env_dir.DefaultConfig, path)
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
