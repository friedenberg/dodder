package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/kilo/queries"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	utility.AddCmd(
		"edit",
		&Edit{
			CheckoutMode: checkout_mode.MetadataOnly,
		})
}

type Edit struct {
	command_components_dodder.LocalWorkingCopyWithQueryGroup

	complete command_components_dodder.Complete

	// TODO-P3 add force
	command_components_dodder.Checkout
	CheckoutMode checkout_mode.Mode
}

var _ interfaces.CommandComponentWriter = (*Edit)(nil)

func (cmd *Edit) SetFlagDefinitions(flagSet interfaces.CLIFlagDefinitions) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagDefinitions(flagSet)

	cmd.Checkout.SetFlagDefinitions(flagSet)

	flagSet.Var(&cmd.CheckoutMode, "mode", "mode for checking out the object")
}

func (cmd Edit) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Tag,
		genres.Zettel,
		genres.Type,
		genres.Repo,
	)
}

func (cmd *Edit) Complete(
	req command.Request,
	envLocal env_local.Env,
	commandLine command.CommandLine,
) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	args := commandLine.FlagsOrArgs[1:]

	if commandLine.InProgress != "" {
		args = args[:len(args)-1]
	}

	cmd.complete.CompleteObjectsIncludingWorkspace(
		req,
		localWorkingCopy,
		queries.BuilderOptionDefaultGenres(genres.Zettel),
		args...,
	)
}

func (cmd Edit) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		queries.BuilderOptions(
			queries.BuilderOptionWorkspace(repo),
			queries.BuilderOptionDefaultGenres(
				genres.Tag,
				genres.Zettel,
				genres.Type,
				genres.Repo,
			),
		),
		repo,
		req.PopArgs(),
	)

	options := checkout_options.Options{
		CheckoutMode: cmd.CheckoutMode,
	}

	opEdit := user_ops.Checkout{
		Repo:            repo,
		Options:         options,
		Edit:            true,
		RefreshCheckout: true,
	}

	if _, err := opEdit.RunQuery(queryGroup); err != nil {
		repo.Cancel(err)
	}
}
