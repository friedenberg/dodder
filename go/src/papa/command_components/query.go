package command_components

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
)

type Query struct {
	sku.ExternalQueryOptions
}

func (cmd *Query) SetFlagSet(f *flag.FlagSet) {
	// TODO switch to repo
	f.Var(&cmd.RepoId, "kasten", "none or Browser")
	f.BoolVar(&cmd.ExcludeUntracked, "exclude-untracked", false, "")
	f.BoolVar(&cmd.ExcludeRecognized, "exclude-recognized", false, "")
}

func (cmd Query) MakeQueryIncludingWorkspace(
	req command.Request,
	options pkg_query.BuilderOption,
	workingCopy repo.WorkingCopy,
	args []string,
) (query *pkg_query.Query) {
	if repo, ok := workingCopy.(repo.LocalWorkingCopy); ok {
		envWorkspace := repo.GetEnvWorkspace()

		options = pkg_query.BuilderOptions(
			options,
			pkg_query.BuilderOptionWorkspace{Env: envWorkspace},
		)
	}

	return cmd.MakeQuery(
		req,
		options,
		workingCopy,
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
		req.CancelWithError(err)
	}

	return
}
