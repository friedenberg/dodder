package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("push", &Push{})
}

type Push struct {
	command_components.LocalWorkingCopy
	command_components.RemoteTransfer
	command_components.Query
}

func (cmd *Push) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.RemoteTransfer.SetFlagSet(flagSet)
	cmd.Query.SetFlagSet(flagSet)
	cmd.LocalWorkingCopy.SetFlagSet(flagSet)
}

func (cmd Push) Run(req command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	var remoteObject *sku.Transacted

	{
		var err error

		if remoteObject, err = localWorkingCopy.GetObjectFromObjectId(
			req.PopArg("repo-id"),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	remote := cmd.MakeRemote(req, localWorkingCopy, remoteObject)

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		query.BuilderOptions(
			query.BuilderOptionDefaultSigil(
				ids.SigilHistory,
				ids.SigilHidden,
			),
			query.BuilderOptionDefaultGenres(genres.InventoryList),
		),
		localWorkingCopy,
		req.PopArgs(),
	)

	if err := remote.(repo.WorkingCopy).PullQueryGroupFromRemote(
		localWorkingCopy,
		queryGroup,
		cmd.WithPrintCopies(true),
	); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
