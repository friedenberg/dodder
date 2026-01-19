package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/env_dir"
	"code.linenisgreat.com/dodder/go/src/juliett/env_local"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
)

// TODO move to command_components
type EnvRepo struct{}

func (cmd EnvRepo) MakeEnvRepo(
	req command.Request,
	permitNoDodderDirectory bool,
) env_repo.Env {
	dir := env_dir.MakeDefault(
		req,
		env_dir.XDGUtilityNameDodder,
		req.Utility.GetConfigDodder().Debug,
	)

	envUI := env_ui.Make(
		req,
		req.Utility.GetConfigDodder(),
		env_ui.Options{},
	)

	var envRepo env_repo.Env

	envRepoOptions := env_repo.Options{
		BasePath:                req.Utility.GetConfigDodder().BasePath,
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
