package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type Genesis struct {
	env_repo.BigBang
	LocalWorkingCopy
	LocalArchive
}

func (cmd *Genesis) SetFlagSet(f *flag.FlagSet) {
	cmd.BigBang.SetFlagSet(f)
}

func (cmd Genesis) OnTheFirstDay(
	req command.Request,
	repoIdString string,
) repo.LocalRepo {
	ui := env_ui.Make(
		req,
		req.Blob,
		env_ui.Options{},
	)

	var repoId ids.RepoId

	if err := repoId.Set(repoIdString); err != nil {
		ui.Cancel(err)
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
			env_local.Make(ui, dir),
			options,
		); err != nil {
			ui.Cancel(err)
		}
	}

	envRepo.Genesis(cmd.BigBang)

	switch cmd.BigBang.GenesisConfig.Blob.GetRepoType() {
	case repo_type.TypeWorkingCopy:
		return local_working_copy.Genesis(
			cmd.BigBang,
			envRepo,
		)

	case repo_type.TypeArchive:
		return cmd.MakeLocalArchive(envRepo)

	default:
		req.Cancel(
			repo_type.ErrUnsupportedRepoType{
				Actual: cmd.BigBang.GenesisConfig.Blob.GetRepoType(),
			})

	}

	return nil
}
