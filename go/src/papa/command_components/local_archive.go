package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

// TODO remove and remove archive repos too
type LocalArchive struct {
	EnvRepo
}

func (cmd *LocalArchive) SetFlagSet(f *flag.FlagSet) {
}

func (cmd LocalArchive) MakeLocalArchive(
	envRepo env_repo.Env,
) *local_working_copy.Repo {
	return local_working_copy.MakeWithEnvRepo(
		local_working_copy.OptionsEmpty,
		envRepo,
	)
}
