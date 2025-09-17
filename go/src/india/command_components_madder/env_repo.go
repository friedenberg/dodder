package command_components_madder

import (
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type EnvRepo struct{}

func (cmd EnvRepo) MakeEnvRepo(
	req command.Request,
	permitNoDodderDirectory bool,
) env_repo.Env {
	dir := env_dir.MakeDefault(
		req,
		req.Config.Debug,
	)

	envUI := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	var envRepo env_repo.Env

	envRepoOptions := env_repo.Options{
		BasePath:                req.Config.BasePath,
		PermitNoDodderDirectory: permitNoDodderDirectory,
	}

	{
		var err error

		if envRepo, err = env_repo.Make(
			env_local.Make(envUI, dir),
			envRepoOptions,
		); err != nil {
			envUI.Cancel(err)
		}
	}

	return envRepo
}

func (cmd EnvRepo) MakeEnvRepoFromEnvLocal(
	envLocal env_local.Env,
) env_repo.Env {
	var envRepo env_repo.Env

	layoutOptions := env_repo.Options{
		BasePath: envLocal.GetCLIConfig().BasePath,
	}

	{
		var err error

		if envRepo, err = env_repo.Make(
			envLocal,
			layoutOptions,
		); err != nil {
			envLocal.Cancel(err)
		}
	}

	return envRepo
}

// TODO move to command_components
