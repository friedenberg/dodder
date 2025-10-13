package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd("reindex", &Reindex{})
}

type Reindex struct {
	command_components_dodder.LocalWorkingCopy
}

func (cmd Reindex) Run(req command.Request) {
	args := req.PopArgs()

	if len(args) > 0 {
		errors.ContextCancelWithErrorf(
			req,
			"reindex does not support arguments",
		)
	}

	localWorkingCopy := cmd.MakeLocalWorkingCopyWithOptions(
		req,
		env_ui.Options{},
		local_working_copy.OptionsAllowConfigReadError,
	)

	localWorkingCopy.Reindex()
}
