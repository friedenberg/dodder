package commands

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/bravo/flags"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	cmd := &Checkin{
		Proto: sku.MakeProto(nil),
	}

	command.Register("checkin", cmd)
	command.Register("add", cmd)
	command.Register("save", cmd)
}

type Checkin struct {
	command_components.LocalWorkingCopyWithQueryGroup

	complete command_components.Complete

	IgnoreBlob bool
	Proto      sku.Proto

	command_components.Checkout

	CheckoutBlobAndRun string
	OpenBlob           bool
}

func (cmd *Checkin) SetFlagSet(flagSet *flags.FlagSet) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagSet(flagSet)

	flagSet.BoolVar(
		&cmd.IgnoreBlob,
		"ignore-blob",
		false,
		"do not change the blob",
	)

	flagSet.StringVar(
		&cmd.CheckoutBlobAndRun,
		"each-blob",
		"",
		"checkout each Blob and run a utility",
	)

	cmd.complete.SetFlagsProto(
		&cmd.Proto,
		flagSet,
		"description to use for new zettels",
		"tags added for new zettels",
		"type used for new zettels",
	)

	cmd.Checkout.SetFlagSet(flagSet)
}

// TODO refactor into common
func (cmd *Checkin) Complete(
	_ command.Request,
	envLocal env_local.Env,
	commandLine command.CommandLine,
) {
	searchDir := envLocal.GetCwd()

	if commandLine.InProgress != "" && files.Exists(commandLine.InProgress) {
		var err error

		if commandLine.InProgress, err = filepath.Abs(commandLine.InProgress); err != nil {
			envLocal.Cancel(err)
			return
		}

		if searchDir, err = filepath.Rel(searchDir, commandLine.InProgress); err != nil {
			envLocal.Cancel(err)
			return
		}
	}

	for dirEntry, err := range files.WalkDir(searchDir) {
		if err != nil {
			envLocal.Cancel(err)
			return
		}

		if files.WalkDirIgnoreFuncHidden(dirEntry) {
			continue
		}

		if !dirEntry.IsDir() {
			envLocal.GetUI().Printf("%s\tfile", dirEntry.RelPath)
		} else {
			envLocal.GetUI().Printf("%s/\tdirectory", dirEntry.RelPath)
		}
	}
}

func (cmd Checkin) Run(dep command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		dep,
		query.BuilderOptions(
			query.BuilderOptionRequireNonEmptyQuery(),
			query.BuilderOptionDefaultSigil(ids.SigilExternal),
			query.BuilderOptionDefaultGenres(genres.All()...),
		),
	)

	workspace := localWorkingCopy.GetEnvWorkspace()
	workspaceTags := workspace.GetDefaults().GetDefaultTags()

	for t := range workspaceTags.All() {
		cmd.Proto.Tags.Add(t)
	}

	op := user_ops.Checkin{
		Delete:             cmd.Delete,
		Organize:           cmd.Organize,
		Proto:              cmd.Proto,
		CheckoutBlobAndRun: cmd.CheckoutBlobAndRun,
		OpenBlob:           cmd.OpenBlob,
	}

	// TODO add auto dot operator
	if err := op.Run(localWorkingCopy, queryGroup); err != nil {
		dep.Cancel(err)
	}
}
