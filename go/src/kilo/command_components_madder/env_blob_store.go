package command_components_madder

import (
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
	"code.linenisgreat.com/dodder/go/src/india/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
)

type EnvBlobStore struct{}

func (cmd EnvBlobStore) MakeEnvBlobStore(
	req command.Request,
) env_repo.BlobStoreEnv {
	dir := env_dir.MakeDefault(
		req,
		req.Utility.GetName(),
		req.Config.Debug,
	)

	envUI := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	envLocal := env_local.Make(envUI, dir)

	return env_repo.MakeBlobStoreEnv(envLocal)
}
