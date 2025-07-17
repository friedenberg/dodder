package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register(
		"edit",
		&Edit{
			CheckoutMode: checkout_mode.MetadataOnly,
		},
	)
}

type Edit struct {
	command_components.LocalWorkingCopyWithQueryGroup

	complete command_components.Complete

	// TODO-P3 add force
	IgnoreWorkspace bool
	command_components.Checkout
	CheckoutMode checkout_mode.Mode
}

func (cmd *Edit) SetFlagSet(flagSet *flag.FlagSet) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagSet(flagSet)

	cmd.Checkout.SetFlagSet(flagSet)

	flagSet.Var(&cmd.CheckoutMode, "mode", "mode for checking out the object")

	flagSet.BoolVar(
		&cmd.IgnoreWorkspace,
		"ignore-workspace",
		false,
		"ignore any workspaces that may be present and checkout the object in a temporary workspace",
	)
}

func (c Edit) CompletionGenres() ids.Genre {
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
		query.BuilderOptionDefaultGenres(genres.Zettel),
		args...,
	)
}

func (cmd Edit) Run(req command.Request) {
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)
	envWorkspace := localWorkingCopy.GetEnvWorkspace()

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		query.BuilderOptions(
			query.BuilderOptionsOld(cmd),
			query.BuilderOptionWorkspace{Env: envWorkspace},
			query.BuilderOptionDefaultGenres(
				genres.Tag,
				genres.Zettel,
				genres.Type,
				genres.Repo,
			),
		),
		localWorkingCopy,
		req.PopArgs(),
	)

	options := checkout_options.Options{
		CheckoutMode: cmd.CheckoutMode,
	}

	opEdit := user_ops.Checkout{
		Repo:            localWorkingCopy,
		Options:         options,
		Edit:            true,
		RefreshCheckout: true,
	}

	if _, err := opEdit.RunQuery(queryGroup); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
