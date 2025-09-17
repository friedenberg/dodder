package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	bigBang := env_repo.BigBang{}
	bigBang.SetDefaults()

	command.Register("init", &Init{
		Genesis: command_components.Genesis{
			BigBang: bigBang,
		},
	})
}

type Init struct {
	command_components.Genesis
}

var _ interfaces.CommandComponentWriter = (*Init)(nil)

func (cmd *Init) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.Genesis.SetFlagDefinitions(flagSet)
}

func (cmd *Init) Run(req command.Request) {
	repoId := req.PopArg("repo-id")
	req.AssertNoMoreArgs()
	cmd.OnTheFirstDay(req, repoId)
}
