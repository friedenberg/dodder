package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("init", &Init{})
}

type Init struct {
	next bool
	command_components.Genesis
}

func (cmd *Init) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.Genesis.SetFlagSet(flagSet)
	flagSet.BoolVar(&cmd.next, "next", false, "use the next store version instead of the current")
}

func (cmd *Init) Run(req command.Request) {
	repoId := req.PopArg("repo-id")

	if cmd.next {
		cmd.Config.StoreVersion = store_version.VNext
	}

	req.AssertNoMoreArgs()
	cmd.OnTheFirstDay(req, repoId)
}
