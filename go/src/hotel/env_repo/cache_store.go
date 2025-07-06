package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

func (env Env) ReadCloserCache(p string) (sha.ReadCloser, error) {
	o := env_dir.FileReadOptions{
		// Config: s.Config.Blob,
		Path: p,
	}

	return env_dir.NewFileReader(o)
}

func (env Env) WriteCloserCache(
	p string,
) (w sha.WriteCloser, err error) {
	return env_dir.NewMover(
		env_dir.MoveOptions{
			// Config:      s.Config.Blob,
			FinalPath:   p,
			TemporaryFS: env.GetTempLocal(),
		},
	)
}
