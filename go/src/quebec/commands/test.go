package commands

import (
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/golf/command"
)

type Test struct{}

func init() {
	command.Register("test", &Test{})
}

func (*Test) SetFlagSet(*flags.FlagSet) {}

func (c Test) Run(dep command.Request) {}
