package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
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
	LocalArchive
}

func (cmd *Genesis) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.BigBang.SetFlagSet(flagSet)
}

func (cmd Genesis) OnTheFirstDay(
	req command.Request,
	repoIdString string,
) *local_working_copy.Repo {
	envUI := env_ui.Make(
		req,
		req.Blob,
		env_ui.Options{},
	)

	var repoId ids.RepoId

	if err := repoId.Set(repoIdString); err != nil {
		envUI.Cancel(err)
	}

	cmd.GenesisConfig.Blob.SetRepoId(repoId)

	dir := env_dir.MakeDefaultAndInitialize(
		req,
		req.Blob.Debug,
		cmd.OverrideXDGWithCwd,
	)

	var envRepo env_repo.Env

	options := env_repo.Options{
		BasePath:                req.Blob.BasePath,
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
	defer ui.Log().Print("genesis done")

	return local_working_copy.Genesis(cmd.BigBang, envRepo)
}
