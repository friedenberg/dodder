package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

func (env Env) ReadCloserCache(path string) (interfaces.ReadCloseDigester, error) {
	return env_dir.NewFileReader(env_dir.DefaultConfig, path)
}

func (env Env) WriteCloserCache(
	path string,
) (interfaces.WriteCloseDigester, error) {
	return env_dir.NewMover(
		env_dir.DefaultConfig,
		env_dir.MoveOptions{
			FinalPath:   path,
			TemporaryFS: env.GetTempLocal(),
		},
	)
}
