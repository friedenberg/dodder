package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
)

type Test struct{}

var _ interfaces.CommandComponentWriter = (*Test)(nil)

func init() {
	command.Register("test", &Test{})
}

func (*Test) SetFlagDefinitions(interfaces.CommandLineFlagDefinitions) {}

func (c Test) Run(dep command.Request) {}
