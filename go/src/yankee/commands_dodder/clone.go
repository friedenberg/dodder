package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
	"code.linenisgreat.com/dodder/go/src/november/queries"
	"code.linenisgreat.com/dodder/go/src/xray/command_components_dodder"
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
	remote, _ := cmd.MakeRemoteAndObject(req, local)

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		queries.BuilderOptions(
			queries.BuilderOptionDefaultSigil(
				ids.SigilHistory,
				ids.SigilHidden,
			),
			queries.BuilderOptionDefaultGenres(genres.InventoryList),
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
