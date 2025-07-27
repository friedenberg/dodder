package commands

import (
	"flag"

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

func (cmd *Init) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.Genesis.SetFlagSet(flagSet)
}

func (cmd *Init) Run(req command.Request) {
	repoId := req.PopArg("repo-id")
	req.AssertNoMoreArgs()
	cmd.OnTheFirstDay(req, repoId)
}
