package command_components_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type RemoteTransfer struct {
	Remote
	repo.ImporterOptions
}

var _ interfaces.CommandComponentWriter = (*RemoteTransfer)(nil)

func (cmd *RemoteTransfer) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
	cmd.Remote.SetFlagDefinitions(flagDefinitions)
	cmd.ImporterOptions.SetFlagDefinitions(flagDefinitions)
}
