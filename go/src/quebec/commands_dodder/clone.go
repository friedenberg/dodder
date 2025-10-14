package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
)

func init() {
	utility.AddCmd(
		"clone",
		&Clone{
			Genesis: command_components_dodder.Genesis{
				BigBang: env_repo.BigBang{
					ExcludeDefaultType: true,
				},
			},
		})
}

type Clone struct {
	command_components_dodder.Genesis
	command_components_dodder.RemoteTransfer
	command_components_dodder.Query
}

var _ interfaces.CommandComponentWriter = (*Clone)(nil)

func (cmd *Clone) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
	cmd.Genesis.SetFlagDefinitions(flagDefinitions)
	cmd.RemoteTransfer.SetFlagDefinitions(flagDefinitions)
	cmd.Query.SetFlagDefinitions(flagDefinitions)
}

func (cmd Clone) Run(req command.Request) {
	local := cmd.OnTheFirstDay(req, req.PopArg("new repo id"))

	// TODO offer option to persist remote object, if supported
	remote, _ := cmd.CreateRemoteObject(req, local)

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		query.BuilderOptions(
			query.BuilderOptionDefaultSigil(
				ids.SigilHistory,
				ids.SigilHidden,
			),
			query.BuilderOptionDefaultGenres(genres.InventoryList),
		),
		local,
		req.PopArgs(),
	)

	if err := local.PullQueryGroupFromRemote(
		remote,
		queryGroup,
		cmd.WithPrintCopies(true),
	); err != nil {
		req.Cancel(err)
	}
}
