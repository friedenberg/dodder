package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

type EnvRepo struct{}

func (cmd *EnvRepo) SetFlagSet(f *flag.FlagSet) {}

func (cmd EnvRepo) MakeEnvRepo(
	dep command.Request,
	permitNoDodderDirectory bool,
) env_repo.Env {
	dir := env_dir.MakeDefault(
		dep,
		dep.Config.Debug,
	)

	ui := env_ui.Make(
		dep,
		dep.Config,
		env_ui.Options{},
	)

	var repoLayout env_repo.Env

	layoutOptions := env_repo.Options{
		BasePath:                dep.Config.BasePath,
		PermitNoDodderDirectory: permitNoDodderDirectory,
	}

	{
		var err error

		if repoLayout, err = env_repo.Make(
			env_local.Make(ui, dir),
			layoutOptions,
		); err != nil {
			ui.CancelWithError(err)
		}
	}

	return repoLayout
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
			envLocal.CancelWithError(err)
		}
	}

	return repoLayout
}

func (EnvRepo) MakeTypedInventoryListBlobStore(
	envRepo env_repo.Env,
) typed_blob_store.InventoryList {
	objectFormat := object_inventory_format.FormatForVersion(
		envRepo.GetStoreVersion(),
	)

	boxFormat := box_format.MakeBoxTransactedArchive(
		envRepo,
		options_print.V0{}.WithPrintTai(true),
	)

	return typed_blob_store.MakeInventoryStore(
		envRepo,
		objectFormat,
		boxFormat,
	)
}
