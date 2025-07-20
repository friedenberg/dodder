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

func (cmd Last) CompletionGenres() ids.Genre {
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

func (cmd Last) runArchive(envRepo env_repo.Env, archive repo.Repo) {
	if (cmd.Edit || cmd.Organize) && cmd.Format.WasSet() {
		errors.ContextCancelWithErrorf(
			envRepo,
			"cannot organize, edit, or specify format for Archive repos",
		)
	}

	// TODO replace with sku.ListFormat
	boxFormat := box_format.MakeBoxTransactedArchive(
		envRepo,
		options_print.Options{}.WithPrintTai(true),
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

	if err := cmd.runWithInventoryList(envRepo, archive, f); err != nil {
		envRepo.Cancel(err)
	}
}

func (cmd Last) runLocalWorkingCopy(localWorkingCopy *local_working_copy.Repo) {
	if (cmd.Edit || cmd.Organize) && cmd.Format.WasSet() {
		ui.Err().Print("ignoring format")
	} else if cmd.Edit && cmd.Organize {
		errors.ContextCancelWithErrorf(localWorkingCopy, "cannot organize and edit at the same time")
	}

	skus := sku.MakeTransactedMutableSet()

	var funcIter interfaces.FuncIter[*sku.Transacted]

	if cmd.Organize || cmd.Edit {
		funcIter = skus.Add
	} else {
		{
			var err error

			if funcIter, err = localWorkingCopy.MakeFormatFunc(
				cmd.Format.String(),
				localWorkingCopy.GetEnvRepo().GetUIFile(),
			); err != nil {
				localWorkingCopy.GetEnvRepo().Cancel(err)
			}
		}
	}

	funcIter = quiter.MakeSyncSerializer(funcIter)

	if err := cmd.runWithInventoryList(
		localWorkingCopy.GetEnvRepo(),
		localWorkingCopy,
		funcIter,
	); err != nil {
		localWorkingCopy.GetEnvRepo().Cancel(err)
	}

	if cmd.Organize {
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
	} else if cmd.Edit {
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

	seq := inventoryListBlobStore.StreamInventoryListBlobSkus(
		listObject,
	)

	for sk, seqError := range seq {
		if seqError != nil {
			ui.Err().Print(seqError)
			continue
		}

		func() {
			// TODO investigate the pool folow for StreamInventoryListBlobSkus and
			// determine why repooling here is breaking things
			// defer sku.GetTransactedPool().Put(sk)

			if err = funcIter(sk); err != nil {
				err = errors.Wrap(err)
				return
			}
		}()
	}

	return
}
