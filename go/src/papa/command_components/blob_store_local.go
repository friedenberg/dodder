package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/foxtrot/config_mutable_cli"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type BlobStoreLocal struct{}

func (cmd *BlobStoreLocal) SetFlagSet(f *flag.FlagSet) {
}

type BlobStoreWithEnv struct {
	env_ui.Env
	interfaces.BlobStore
}

func (c BlobStoreLocal) MakeBlobStoreLocal(
	context errors.Context,
	config config_mutable_cli.Config,
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

	var repoLayout env_repo.Env

	{
		var err error

		if repoLayout, err = env_repo.Make(
			env_local.Make(ui, dir),
			layoutOptions,
		); err != nil {
			context.CancelWithError(err)
		}
	}

	return BlobStoreWithEnv{
		Env:       ui,
		BlobStore: repoLayout,
	}
}
