package command_components

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type RemoteTransfer struct {
	Remote
	repo.ImporterOptions
}

func (cmd *RemoteTransfer) SetFlagSet(
	flagDefinitions interfaces.CommandLineFlagDefinitions,
) {
	cmd.Remote.SetFlagSet(flagDefinitions)
	cmd.ImporterOptions.SetFlagSet(flagDefinitions)
}
