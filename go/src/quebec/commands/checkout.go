package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register(
		"checkout",
		&Checkout{
			CheckoutOptions: checkout_options.Options{
				CheckoutMode: checkout_mode.MetadataOnly,
			},
		},
	)
}

type Checkout struct {
	command_components.LocalWorkingCopyWithQueryGroup

	CheckoutOptions checkout_options.Options
	Organize        bool
}

func (cmd *Checkout) SetFlagSet(f *flag.FlagSet) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagSet(f)
	f.BoolVar(&cmd.Organize, "organize", false, "")
	cmd.CheckoutOptions.SetFlagSet(f)
}


func (cmd Checkout) Run(req command.Request) {
	repo := cmd.MakeLocalWorkingCopy(req)
	envWorkspace := repo.GetEnvWorkspace()

	queryGroup := cmd.MakeQueryIncludingWorkspace(
		req,
		query.BuilderOptions(
			query.BuilderOptionPermittedSigil(ids.SigilLatest),
			query.BuilderOptionPermittedSigil(ids.SigilHidden),
			query.BuilderOptionRequireNonEmptyQuery(),
			query.BuilderOptionWorkspace(repo),
			query.BuilderOptionDefaultGenres(genres.Zettel),
		),
		repo,
		req.PopArgs(),
	)

	opCheckout := user_ops.Checkout{
		Repo:     repo,
		Organize: cmd.Organize,
		Options:  cmd.CheckoutOptions,
	}

	envWorkspace.AssertNotTemporaryOrOfferToCreate(repo)

	if _, err := opCheckout.RunQuery(queryGroup); err != nil {
		repo.Cancel(err)
	}
}
