package command_components_madder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
	"code.linenisgreat.com/dodder/go/src/juliett/env_local"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
)

// TODO remove and replace with BlobStore
type BlobStoreLocal struct{}

var _ interfaces.CommandComponentWriter = (*BlobStoreLocal)(nil)

func (cmd *BlobStoreLocal) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
}

type BlobStoreWithEnv struct {
	env_ui.Env
	interfaces.BlobStore
}

func (cmd BlobStoreLocal) MakeBlobStoreLocal(
	req command.Request,
	config repo_config_cli.Config,
	envOptions env_ui.Options,
) BlobStoreWithEnv {
	dir := env_dir.MakeDefault(
		req,
		req.Utility.GetName(),
		config.Debug,
	)

	ui := env_ui.Make(
		req,
		config,
		envOptions,
	)

	layoutOptions := env_repo.Options{
		BasePath: config.BasePath,
	}

	var envRepo env_repo.Env

	{
		var err error

		if envRepo, err = env_repo.Make(
			env_local.Make(ui, dir),
			layoutOptions,
		); err != nil {
			req.Cancel(err)
		}
	}

	return BlobStoreWithEnv{
		Env:       ui,
		BlobStore: envRepo.GetDefaultBlobStore(),
	}
}
