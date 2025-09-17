package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("show", &Show{
		Format: local_working_copy.FormatFlag{
			DefaultFormatter: local_working_copy.GetFormatFuncConstructorEntry(
				"log",
			),
		},
	})
}

type Show struct {
	command_components_madder.EnvRepo
	command_components.LocalWorkingCopy
	command_components.Query
	command_components.RemoteTransfer

	complete command_components.Complete

	After      ids.Tai
	Before     ids.Tai
	Format     local_working_copy.FormatFlag
	RemoteRepo ids.RepoId
}

var _ interfaces.CommandComponentWriter = (*Show)(nil)

func (cmd *Show) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagDefinitions(flagSet)
	cmd.Query.SetFlagDefinitions(flagSet)

	flagSet.Var(
		&cmd.Format,
		"format",
		"format used when outputting objects to stdout",
	)
	flagSet.Var((*ids.TaiRFC3339Value)(&cmd.Before), "before", "")
	flagSet.Var((*ids.TaiRFC3339Value)(&cmd.After), "after", "")
	flagSet.Var(&cmd.RemoteRepo, "repo", "the remote repo to query")
}

func (cmd Show) Complete(
	req command.Request,
	envLocal env_local.Env,
	commandLine command.CommandLine,
) {
	repo := cmd.MakeLocalWorkingCopy(req)

	args := commandLine.FlagsOrArgs[1:]

	if commandLine.InProgress != "" {
		args = args[:len(args)-1]
	}

	cmd.complete.CompleteObjects(
		req,
		repo,
		pkg_query.BuilderOptionDefaultGenres(genres.Tag),
		args...,
	)
}

func (cmd Show) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	args := req.PopArgs()

	query := cmd.MakeQueryIncludingWorkspace(
		req,
		pkg_query.BuilderOptions(
			pkg_query.BuilderOptionWorkspace(repo),
			pkg_query.BuilderOptionDefaultGenres(genres.Zettel),
		),
		repo,
		args,
	)

	cmd.runWithLocalWorkingCopyAndQuery(req, repo, query)
}

func (cmd Show) runWithLocalWorkingCopyAndQuery(
	req command.Request,
	localWorkingCopy *local_working_copy.Repo,
	query *pkg_query.Query,
) {
	var remoteObject *sku.Transacted
	var remoteWorkingCopy repo.Repo

	if !cmd.RemoteRepo.IsEmpty() {
		var err error

		if remoteObject, err = localWorkingCopy.GetObjectFromObjectId(
			cmd.RemoteRepo.StringWithSlashPrefix(),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}

		remoteRepo := cmd.MakeRemote(req, localWorkingCopy, remoteObject)
		remoteWorkingCopy, _ = remoteRepo.(repo.Repo)
	}

	if cmd.Format.GetName() == "" && pkg_query.IsExactlyOneObjectId(query) {
		// if err := cmd.Format.Set("text"); err != nil {
		// 	localWorkingCopy.Cancel(err)
		// 	return
		// }
	}

	output := cmd.Format.MakeFormatFunc(
		localWorkingCopy,
		localWorkingCopy.GetUIFile(),
	)

	if !cmd.Before.IsEmpty() {
		old := output

		output = func(sk *sku.Transacted) (err error) {
			if !sk.GetTai().Before(cmd.Before) {
				return
			}

			if err = old(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}
	}

	if !cmd.After.IsEmpty() {
		old := output

		output = func(sk *sku.Transacted) (err error) {
			if !sk.GetTai().After(cmd.After) {
				return
			}

			if err = old(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		}
	}

	if remoteWorkingCopy != nil {
		var list *sku.ListTransacted

		{
			var err error

			if list, err = remoteWorkingCopy.MakeInventoryList(query); err != nil {
				localWorkingCopy.Cancel(err)
			}
		}

		for sk := range list.All() {
			if err := output(sk); err != nil {
				localWorkingCopy.Cancel(err)
			}
		}
	} else {
		if err := localWorkingCopy.GetStore().QueryTransacted(
			query,
			quiter.MakeSyncSerializer(output),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}
}
