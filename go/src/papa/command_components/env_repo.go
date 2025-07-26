package command_components

import (
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
)

type EnvRepo struct{}

func (cmd EnvRepo) MakeEnvRepo(
	dep command.Request,
	permitNoDodderDirectory bool,
) env_repo.Env {
	dir := env_dir.MakeDefault(
		dep,
		dep.Blob.Debug,
	)

	ui := env_ui.Make(
		dep,
		dep.Blob,
		env_ui.Options{},
	)

	var envRepo env_repo.Env

	envRepoOptions := env_repo.Options{
		BasePath:                dep.Blob.BasePath,
		PermitNoDodderDirectory: permitNoDodderDirectory,
	}

	{
		var err error

		if envRepo, err = env_repo.Make(
			env_local.Make(ui, dir),
			envRepoOptions,
		); err != nil {
			ui.Cancel(err)
		}
	}

	return envRepo
}

func (cmd EnvRepo) MakeEnvRepoFromEnvLocal(
	envLocal env_local.Env,
) env_repo.Env {
	var repoLayout env_repo.Env

	layoutOptions := env_repo.Options{
		BasePath: envLocal.GetCLIConfig().BasePath,
	}

	{
		var err error

		if repoLayout, err = env_repo.Make(
			envLocal,
			layoutOptions,
		); err != nil {
			envLocal.Cancel(err)
		}
	}

	return repoLayout
}

func (EnvRepo) MakeTypedInventoryListBlobStore(
	envRepo env_repo.Env,
) inventory_list_coders.Closet {
	boxFormat := box_format.MakeBoxTransactedArchive(
		envRepo,
		options_print.Options{}.WithPrintTai(true),
	)

	return inventory_list_coders.MakeCloset(
		envRepo,
		boxFormat,
	)
}
