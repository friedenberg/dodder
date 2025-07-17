package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/bravo/values"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("last", &Last{
		Format: values.MakeStringDefault("log"),
	})
}

type Last struct {
	command_components.EnvRepo
	command_components.LocalArchive

	RepoId   ids.RepoId
	Edit     bool
	Organize bool
	Format   values.String
}

func (cmd *Last) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.LocalArchive.SetFlagSet(flagSet)

	flagSet.Var(&cmd.RepoId, "kasten", "none or Browser")
	flagSet.Var(&cmd.Format, "format", "format")
	flagSet.BoolVar(&cmd.Organize, "organize", false, "")
	flagSet.BoolVar(&cmd.Edit, "edit", false, "")
}

func (c Last) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.InventoryList,
	)
}

func (cmd Last) Run(dep command.Request) {
	repoLayout := cmd.MakeEnvRepo(dep, false)

	archive := cmd.MakeLocalArchive(repoLayout)

	if len(dep.PopArgs()) != 0 {
		ui.Err().Print("ignoring arguments")
	}

	if localWorkingCopy, ok := archive.(*local_working_copy.Repo); ok {
		cmd.runLocalWorkingCopy(localWorkingCopy)
	} else {
		cmd.runArchive(repoLayout, archive)
	}
}

func (c Last) runArchive(envRepo env_repo.Env, archive repo.Repo) {
	if (c.Edit || c.Organize) && c.Format.WasSet() {
		errors.ContextCancelWithErrorf(
			envRepo,
			"cannot organize, edit, or specify format for Archive repos",
		)
	}

	// TODO replace with sku.ListFormat
	boxFormat := box_format.MakeBoxTransactedArchive(
		envRepo,
		options_print.V0{}.WithPrintTai(true),
	)

	f := string_format_writer.MakeDelim(
		"\n",
		envRepo.GetUIFile(),
		string_format_writer.MakeFunc(
			func(w interfaces.WriterAndStringWriter, o *sku.Transacted) (n int64, err error) {
				return boxFormat.EncodeStringTo(o, w)
			},
		),
	)

	f = quiter.MakeSyncSerializer(f)

	if err := c.runWithInventoryList(envRepo, archive, f); err != nil {
		envRepo.Cancel(err)
	}
}

func (c Last) runLocalWorkingCopy(localWorkingCopy *local_working_copy.Repo) {
	if (c.Edit || c.Organize) && c.Format.WasSet() {
		ui.Err().Print("ignoring format")
	} else if c.Edit && c.Organize {
		errors.ContextCancelWithErrorf(localWorkingCopy, "cannot organize and edit at the same time")
	}

	skus := sku.MakeTransactedMutableSet()

	var funcIter interfaces.FuncIter[*sku.Transacted]

	if c.Organize || c.Edit {
		funcIter = skus.Add
	} else {
		{
			var err error

			if funcIter, err = localWorkingCopy.MakeFormatFunc(
				c.Format.String(),
				localWorkingCopy.GetEnvRepo().GetUIFile(),
			); err != nil {
				localWorkingCopy.GetEnvRepo().Cancel(err)
			}
		}
	}

	funcIter = quiter.MakeSyncSerializer(funcIter)

	if err := c.runWithInventoryList(
		localWorkingCopy.GetEnvRepo(),
		localWorkingCopy,
		funcIter,
	); err != nil {
		localWorkingCopy.GetEnvRepo().Cancel(err)
	}

	if c.Organize {
		opOrganize := user_ops.Organize{
			Repo: localWorkingCopy,
			Metadata: organize_text.Metadata{
				OptionCommentSet: organize_text.MakeOptionCommentSet(nil),
			},
		}

		var results organize_text.OrganizeResults

		{
			var err error

			if results, err = opOrganize.RunWithTransacted(nil, skus); err != nil {
				localWorkingCopy.GetEnvRepo().Cancel(err)
			}
		}

		if _, err := localWorkingCopy.LockAndCommitOrganizeResults(results); err != nil {
			localWorkingCopy.GetEnvRepo().Cancel(err)
		}
	} else if c.Edit {
		opCheckout := user_ops.Checkout{
			Options: checkout_options.Options{
				CheckoutMode: checkout_mode.MetadataAndBlob,
			},
			Repo: localWorkingCopy,
			Edit: true,
		}

		if _, err := opCheckout.Run(skus); err != nil {
			localWorkingCopy.GetEnvRepo().Cancel(err)
		}
	}
}

func (cmd Last) runWithInventoryList(
	envRepo env_repo.Env,
	repo repo.Repo,
	funcIter interfaces.FuncIter[*sku.Transacted],
) (err error) {
	var listObject *sku.Transacted

	inventoryListStore := repo.GetInventoryListStore()

	if listObject, err = inventoryListStore.ReadLast(); err != nil {
		err = errors.Wrap(err)
		return
	}

	inventoryListBlobStore := cmd.MakeTypedInventoryListBlobStore(
		envRepo,
	)

	var listWithBlob sku.TransactedWithBlob[*sku.List]

	if listWithBlob, _, err = inventoryListBlobStore.GetTransactedWithBlob(
		listObject,
	); err != nil {
		err = errors.Wrapf(err, "InventoryList: %q", sku.String(listObject))
		return
	}

	ui.TodoP3("support log line format for skus")
	for sk := range listWithBlob.Blob.All() {
		if err = funcIter(sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
