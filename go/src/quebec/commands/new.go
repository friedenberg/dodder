package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/script_value"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/lima/store_fs"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register("new", &New{})
}

type New struct {
	command_components.LocalWorkingCopy

	complete command_components.Complete

	ids.RepoId
	Count int
	// TODO combine organize and edit and refactor
	command_components.Checkout
	PrintOnly bool
	Filter    script_value.ScriptValue
	Shas      bool

	sku.Proto
}

var _ interfaces.CommandComponentWriter = (*New)(nil)

func (cmd *New) SetFlagDefinitions(flagSet interfaces.CommandLineFlagDefinitions) {
	flagSet.Var(&cmd.RepoId, "kasten", "none or Browser")

	flagSet.BoolVar(
		&cmd.Shas,
		"shas",
		false,
		"treat arguments as blobs that are already checked in",
	)

	flagSet.IntVar(
		&cmd.Count,
		"count",
		1,
		"when creating new empty zettels, how many to create. otherwise ignored",
	)

	flagSet.Var(
		&cmd.Filter,
		"filter",
		"a script to run for each file to transform it the standard zettel format",
	)

	cmd.complete.SetFlagsProto(
		&cmd.Proto,
		flagSet,
		"description to use for new zettels",
		"tags added for new zettels",
		"type used for new zettels",
	)

	cmd.Checkout.SetFlagDefinitions(flagSet)
}

func (cmd New) ValidateFlagsAndArgs(
	repo *local_working_copy.Repo,
	args ...string,
) (err error) {
	if repo.GetConfig().IsDryRun() && len(args) == 0 {
		err = errors.ErrorWithStackf(
			"when -dry-run is set, paths to existing zettels must be provided",
		)
		return err
	}

	return err
}

func (cmd *New) Run(req command.Request) {
	args := req.PopArgs()
	repo := cmd.MakeLocalWorkingCopy(req)

	if err := cmd.ValidateFlagsAndArgs(repo, args...); err != nil {
		repo.Cancel(err)
	}

	cotfo := checkout_options.TextFormatterOptions{}

	f := object_metadata.MakeTextFormat(
		object_metadata.Dependencies{
			EnvDir:    repo.GetEnvRepo(),
			BlobStore: repo.GetEnvRepo().GetDefaultBlobStore(),
		},
	)

	var objects sku.TransactedMutableSet

	if len(args) == 0 {
		emptyOp := user_ops.WriteNewZettels{
			Repo: repo,
		}

		{
			var err error

			if objects, err = emptyOp.RunMany(cmd.Proto, cmd.Count); err != nil {
				repo.Cancel(err)
			}
		}
	} else if cmd.Shas {
		opCreateFromShas := user_ops.CreateFromShas{
			Repo:  repo,
			Proto: cmd.Proto,
		}

		{
			var err error

			if objects, err = opCreateFromShas.Run(args...); err != nil {
				repo.Cancel(err)
			}
		}
	} else {
		opCreateFromPath := user_ops.CreateFromPaths{
			Repo:       repo,
			TextParser: f,
			Filter:     cmd.Filter,
			Delete:     cmd.Delete,
			Proto:      cmd.Proto,
		}

		{
			var err error

			if objects, err = opCreateFromPath.Run(args...); err != nil {
				if errors.IsNotExist(err) {
					errors.ContextCancelWithBadRequestf(repo, "Expected a valid file path. Did you mean to add `-description`?")
				} else {
					repo.Cancel(err)
				}
			}
		}
	}

	// TODO make mutually exclusive with organize
	if cmd.Edit {
		opCheckout := user_ops.Checkout{
			Repo: repo,
			Options: checkout_options.Options{
				CheckoutMode: checkout_mode.MetadataAndBlob,
				OptionsWithoutMode: checkout_options.OptionsWithoutMode{
					StoreSpecificOptions: store_fs.CheckoutOptions{
						ForceInlineBlob:      true,
						TextFormatterOptions: cotfo,
					},
				},
			},
			Edit:            true,
			RefreshCheckout: true,
		}

		if _, err := opCheckout.Run(objects); err != nil {
			repo.Cancel(err)
		}
	}

	if cmd.Organize {
		opOrganize := user_ops.Organize{
			Repo: repo,
		}

		if err := opOrganize.Metadata.SetFromObjectMetadata(
			&cmd.Metadata,
			ids.RepoId{},
		); err != nil {
			repo.Cancel(err)
		}

		var results organize_text.OrganizeResults

		{
			var err error

			if results, err = opOrganize.RunWithTransacted(nil, objects); err != nil {
				repo.Cancel(err)
			}
		}

		if _, err := repo.LockAndCommitOrganizeResults(
			results,
		); err != nil {
			repo.Cancel(err)
		}
	}
}
