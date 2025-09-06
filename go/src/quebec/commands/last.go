package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/lima/repo"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("last", &Last{
		Format: local_working_copy.FormatFlag{
			DefaultFormatter: local_working_copy.GetFormatFuncConstructorEntry(
				"log",
			),
		},
	})
}

type Last struct {
	command_components.EnvRepo
	command_components.LocalWorkingCopy

	RepoId   ids.RepoId
	Edit     bool
	Organize bool
	Format   local_working_copy.FormatFlag
}

func (cmd *Last) SetFlagSet(flagSet interfaces.CommandLineFlagDefinitions) {
	cmd.LocalWorkingCopy.SetFlagSet(flagSet)

	// TODO remove
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

func (cmd Last) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	if len(req.PopArgs()) != 0 {
		ui.Err().Print("ignoring arguments")
	}

	cmd.runLocalWorkingCopy(repo)
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
		funcIter = cmd.Format.MakeFormatFunc(
			localWorkingCopy,
			localWorkingCopy.GetUIFile(),
		)
	}

	funcIter = quiter.MakeSyncSerializer(funcIter)

	if err := cmd.runWithInventoryList(
		localWorkingCopy.GetEnvRepo(),
		localWorkingCopy,
		funcIter,
	); err != nil {
		localWorkingCopy.GetEnvRepo().Cancel(err)
		return
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

	inventoryListBlobStore := cmd.MakeInventoryListCoderCloset(
		envRepo,
	)

	seq := inventoryListBlobStore.StreamInventoryListBlobSkus(
		listObject,
	)

	for object, seqError := range seq {
		if seqError != nil {
			ui.CLIErrorTreeEncoder.EncodeTo(seqError, ui.Err())
			continue
		}

		func() {
			// TODO investigate the pool folow for StreamInventoryListBlobSkus
			// and
			// determine why repooling here is breaking things
			// defer sku.GetTransactedPool().Put(sk)

			if err = funcIter(object); err != nil {
				err = errors.Wrap(err)
				return
			}
		}()
	}

	return
}
