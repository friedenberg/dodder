package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

func (env Env) ReadCloserCache(path string) (sha.ReadCloser, error) {
	options := env_dir.FileReadOptions{
		Path: path,
	}

	return env_dir.NewFileReader(options)
}

func (env Env) WriteCloserCache(
	path string,
) (sha.WriteCloser, error) {
	return env_dir.NewMover(
		env_dir.MoveOptions{
			FinalPath:   path,
			TemporaryFS: env.GetTempLocal(),
		},
	)
}
