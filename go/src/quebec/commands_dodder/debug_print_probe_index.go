//go:build debug

package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

type DebugPrintProbeIndex struct {
	command_components_dodder.LocalWorkingCopy
}

var _ interfaces.CommandComponentWriter = (*DebugPrintProbeIndex)(nil)

func init() {
	utility.AddCmd("debug-print-probe-index", &DebugPrintProbeIndex{})
}

func (*DebugPrintProbeIndex) SetFlagDefinitions(
	interfaces.CLIFlagDefinitions,
) {
}

func (cmd DebugPrintProbeIndex) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	if err := repo.GetStore().GetStreamIndex().PrintAllProbes(); err != nil {
		repo.Cancel(err)
	}
}
