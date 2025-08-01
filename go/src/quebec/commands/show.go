package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
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
	command_components.EnvRepo
	command_components.LocalArchive
	command_components.Query
	command_components.RemoteTransfer

	complete command_components.Complete

	After      ids.Tai
	Before     ids.Tai
	Format     local_working_copy.FormatFlag
	RemoteRepo ids.RepoId
}

func (cmd *Show) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.LocalArchive.SetFlagSet(flagSet)
	cmd.Query.SetFlagSet(flagSet)

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
	envRepo := cmd.MakeEnvRepo(req, false)
	repo := cmd.MakeLocalArchive(envRepo)

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
	envRepo := cmd.MakeEnvRepo(req, false)
	repo := cmd.MakeLocalArchive(envRepo)

	args := req.PopArgs()

	query := cmd.MakeQueryIncludingWorkspace(
		req,
		pkg_query.BuilderOptions(
			pkg_query.BuilderOptionWorkspace{
				Env: repo.GetEnvWorkspace(),
			},
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
	var remoteWorkingCopy repo.WorkingCopy

	if !cmd.RemoteRepo.IsEmpty() {
		var err error

		if remoteObject, err = localWorkingCopy.GetObjectFromObjectId(
			cmd.RemoteRepo.StringWithSlashPrefix(),
		); err != nil {
			localWorkingCopy.Cancel(err)
		}

		remoteRepo := cmd.MakeRemote(req, localWorkingCopy, remoteObject)
		remoteWorkingCopy, _ = remoteRepo.(repo.WorkingCopy)
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

// TODO add support for query group
func (cmd Show) runWithArchive(
	env env_ui.Env,
	archive repo.Repo,
) {
	// TODO replace with sku.ListFormat
	boxFormat := box_format.MakeBoxTransactedArchive(
		env,
		env.GetCLIConfig().PrintOptions,
	)

	printer := string_format_writer.MakeDelim(
		"\n",
		env.GetUIFile(),
		string_format_writer.MakeFunc(
			func(
				writer interfaces.WriterAndStringWriter,
				object *sku.Transacted,
			) (n int64, err error) {
				errors.ContextContinueOrPanic(env)
				return boxFormat.EncodeStringTo(object, writer)
			},
		),
	)

	inventoryListStore := archive.GetInventoryListStore()

	if err := inventoryListStore.ReadAllSkus(
		func(_, sk *sku.Transacted) (err error) {
			if err = printer(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		env.Cancel(err)
	}
}
