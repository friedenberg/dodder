package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

func init() {
	command.Register(
		"init-archive",
		&InitArchive{
			BigBang: env_repo.BigBang{},
		},
	)
}

type InitArchive struct {
	env_repo.BigBang
}

func (cmd *InitArchive) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.BigBang.SetFlagSet(flagSet)
	cmd.GenesisConfig.Blob.SetRepoType(repo_type.TypeArchive)
}

func (cmd InitArchive) Run(req command.Request) {
	dir := env_dir.MakeDefaultAndInitialize(
		req,
		req.Config.Debug,
		cmd.OverrideXDGWithCwd,
	)

	ui := env_ui.Make(
		req,
		req.Config,
		env_ui.Options{},
	)

	var envRepo env_repo.Env

	layoutOptions := env_repo.Options{
		BasePath:                req.Config.BasePath,
		PermitNoDodderDirectory: true,
	}

	{
		var err error

		if envRepo, err = env_repo.Make(
			env_local.Make(ui, dir),
			layoutOptions,
		); err != nil {
			ui.Cancel(err)
		}
	}

	envRepo.Genesis(cmd.BigBang)
}
