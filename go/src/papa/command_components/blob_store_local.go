package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type BlobStoreLocal struct{}

func (cmd *BlobStoreLocal) SetFlagSet(flagSet *flag.FlagSet) {
}

type BlobStoreWithEnv struct {
	env_ui.Env
	interfaces.BlobStore
}

func (cmd BlobStoreLocal) MakeBlobStoreLocal(
	context interfaces.Context,
	config repo_config_cli.Config,
	envOptions env_ui.Options,
	repoOptions local_working_copy.Options,
) BlobStoreWithEnv {
	dir := env_dir.MakeDefault(
		context,
		config.Debug,
	)

	ui := env_ui.Make(
		context,
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
			context.Cancel(err)
		}
	}

	return BlobStoreWithEnv{
		Env:       ui,
		BlobStore: envRepo.GetDefaultBlobStore(),
	}
}
