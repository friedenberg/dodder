package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"clone",
		&Clone{
			Genesis: command_components.Genesis{
				BigBang: env_repo.BigBang{
					ExcludeDefaultType: true,
				},
			},
		},
	)
}

type Clone struct {
	command_components.Genesis
	command_components.RemoteTransfer
	command_components.Query
}

func (cmd *Clone) SetFlagSet(flagSet *flags.FlagSet) {
	cmd.Genesis.SetFlagSet(flagSet)
	cmd.RemoteTransfer.SetFlagSet(flagSet)
	cmd.Query.SetFlagSet(flagSet)

	// must happen after genesis set flag set as cmd.Config is nil until then
	cmd.GenesisConfig.Blob.SetRepoType(repo_type.TypeWorkingCopy)
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
