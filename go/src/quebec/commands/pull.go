package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("pull", &Pull{})
}

type Pull struct {
	command_components.LocalWorkingCopy
	command_components.RemoteTransfer
	command_components.Query
}

func (cmd *Pull) SetFlagSet(f *flag.FlagSet) {
	cmd.RemoteTransfer.SetFlagSet(f)
	cmd.Query.SetFlagSet(f)
	cmd.LocalWorkingCopy.SetFlagSet(f)
}

func (cmd Pull) Run(req command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	var object *sku.Transacted

	{
		var err error

		if object, err = localWorkingCopy.GetObjectFromObjectId(
			req.PopArg("repo-id"),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	remote := cmd.MakeRemote(req, localWorkingCopy, object)

	qg := cmd.MakeQueryIncludingWorkspace(
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

	switch remote := remote.(type) {
	case repo.WorkingCopy:
		if err := localWorkingCopy.PullQueryGroupFromRemote(
			remote,
			qg,
			cmd.WithPrintCopies(true),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}

	case repo.Repo:
		errors.ContextCancelWithBadRequestf(
			localWorkingCopy,
			"unsupported repo type: %s (%T)",
			remote.GetImmutableConfigPublic().GetRepoType(),
			remote,
		)
	}
}
