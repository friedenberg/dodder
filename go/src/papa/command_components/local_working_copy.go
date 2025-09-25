package command_components

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type LocalWorkingCopy struct {
	Env
}

var _ interfaces.CommandComponentWriter = (*LocalWorkingCopy)(nil)

func (cmd *LocalWorkingCopy) SetFlagDefinitions(f interfaces.CommandLineFlagDefinitions) {
}

func (cmd LocalWorkingCopy) MakeLocalWorkingCopy(
	req command.Request,
) *local_working_copy.Repo {
	return cmd.MakeLocalWorkingCopyWithOptions(
		req,
		env_ui.Options{},
		local_working_copy.OptionsEmpty,
	)
}

func (cmd LocalWorkingCopy) MakeLocalWorkingCopyWithOptions(
	req command.Request,
	envOptions env_ui.Options,
	repoOptions local_working_copy.Options,
) *local_working_copy.Repo {
	env := cmd.MakeEnvWithOptions(req, envOptions)

	return local_working_copy.Make(env, repoOptions)
}

// TODO modify to work with archives
func (cmd LocalWorkingCopy) MakeLocalWorkingCopyFromConfigAndXDGDotenvPath(
	req command.Request,
	xdgDotenvPath string,
	options env_ui.Options,
) (local *local_working_copy.Repo) {
	envLocal := cmd.MakeEnvWithXDGLayoutAndOptions(
		req,
		xdgDotenvPath,
		options,
	)

	local = local_working_copy.Make(
		envLocal,
		local_working_copy.OptionsEmpty,
	)

	return local
}

func (cmd LocalWorkingCopy) MakeLocalWorkingCopyFromEnvLocal(
	envLocal env_local.Env,
) (local *local_working_copy.Repo) {
	local = local_working_copy.Make(
		envLocal,
		local_working_copy.OptionsEmpty,
	)

	return local
}
