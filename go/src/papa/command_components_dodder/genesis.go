package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type Genesis struct {
	env_repo.BigBang
	LocalWorkingCopy
}

var _ interfaces.CommandComponentWriter = (*Genesis)(nil)

func (cmd *Genesis) SetFlagDefinitions(
	flagSet interfaces.CLIFlagDefinitions,
) {
	cmd.BigBang.SetFlagDefinitions(flagSet)
}

func (cmd Genesis) OnTheFirstDay(
	req command.Request,
	repoIdString string,
) *local_working_copy.Repo {
	envUI := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	var repoId ids.RepoId

	if err := repoId.Set(repoIdString); err != nil {
		envUI.Cancel(err)
	}

	cmd.GenesisConfig.Blob.SetRepoId(repoId)

	dir := env_dir.MakeDefaultAndInitialize(
		req,
		env_dir.XDGUtilityNameDodder,
		req.Config.Debug,
		cmd.OverrideXDGWithCwd,
	)

	var envRepo env_repo.Env

	options := env_repo.Options{
		BasePath:                req.Config.BasePath,
		PermitNoDodderDirectory: true,
	}

	{
		var err error

		if envRepo, err = env_repo.Make(
			env_local.Make(envUI, dir),
			options,
		); err != nil {
			envUI.Cancel(err)
		}
	}

	envRepo.Genesis(cmd.BigBang)

	return local_working_copy.Genesis(cmd.BigBang, envRepo)
}
