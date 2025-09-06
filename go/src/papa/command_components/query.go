package command_components

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
)

type Query struct {
	sku.ExternalQueryOptions
}

func (cmd *Query) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	// TODO switch to repo
	flagSet.Var(&cmd.RepoId, "kasten", "none or Browser")
	flagSet.BoolVar(&cmd.ExcludeUntracked, "exclude-untracked", false, "")
	flagSet.BoolVar(&cmd.ExcludeRecognized, "exclude-recognized", false, "")
}

func (cmd Query) MakeQueryIncludingWorkspace(
	req command.Request,
	options pkg_query.BuilderOption,
	repo *local_working_copy.Repo,
	args []string,
) (query *pkg_query.Query) {
	options = pkg_query.BuilderOptions(
		options,
		pkg_query.BuilderOptionWorkspace(repo),
	)

	return cmd.MakeQuery(
		req,
		options,
		repo,
		args,
	)
}

func (cmd Query) MakeQuery(
	req command.Request,
	options pkg_query.BuilderOption,
	workingCopy repo.WorkingCopy,
	args []string,
) (query *pkg_query.Query) {
	var err error

	if query, err = workingCopy.MakeExternalQueryGroup(
		options,
		cmd.ExternalQueryOptions,
		args...,
	); err != nil {
		req.Cancel(err)
	}

	return
}
