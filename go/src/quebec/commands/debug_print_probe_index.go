//go:build debug

package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

type DebugPrintProbeIndex struct {
	command_components.LocalWorkingCopy
}

var _ interfaces.CommandComponentWriter = (*DebugPrintProbeIndex)(nil)

func init() {
	command.Register("debug-print-probe-index", &DebugPrintProbeIndex{})
}

func (*DebugPrintProbeIndex) SetFlagDefinitions(
	interfaces.CommandLineFlagDefinitions,
) {
}

func (cmd DebugPrintProbeIndex) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	if err := repo.GetStore().GetStreamIndex().PrintAllProbes(); err != nil {
		repo.Cancel(err)
	}
}
