package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
)

type Test struct{}

func init() {
	command.Register("test", &Test{})
}

func (*Test) SetFlagSet(interfaces.CommandLineFlagDefinitions) {}

func (c Test) Run(dep command.Request) {}
