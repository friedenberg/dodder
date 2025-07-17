package commands

import (
	"flag"
	"fmt"
	"path"

	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("deinit", &Deinit{})
}

type Deinit struct {
	command_components.LocalWorkingCopy

	Force bool
}

func (cmd *Deinit) SetFlagSet(f *flag.FlagSet) {
	f.BoolVar(
		&cmd.Force,
		"force",
		false,
		"force deinit",
	)
}

func (cmd Deinit) Run(dep command.Request) {
	// TODO switch to archive
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	if !cmd.Force && !cmd.getPermission(localWorkingCopy) {
		ui.Err().Print("permission denied and -force not specified, aborting")
		return
	}

	base := path.Join(localWorkingCopy.GetEnvRepo().Dir())

	if err := files.SetAllowUserChangesRecursive(base); err != nil {
		localWorkingCopy.Cancel(err)
	}

	if err := localWorkingCopy.GetEnvRepo().Delete(
		localWorkingCopy.GetEnvRepo().GetXDG().GetXDGPaths()...,
	); err != nil {
		localWorkingCopy.Cancel(err)
	}

	if err := localWorkingCopy.GetEnvWorkspace().DeleteWorkspace(); err != nil {
		localWorkingCopy.Cancel(err)
	}
}

func (cmd Deinit) getPermission(repo *local_working_copy.Repo) bool {
	return repo.Confirm(
		fmt.Sprintf(
			"are you sure you want to deinit in %q?",
			repo.GetEnvRepo().Dir(),
		),
	)
}
