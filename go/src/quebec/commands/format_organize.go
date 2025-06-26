package commands

import (
	"flag"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	command.Register(
		"format-organize",
		&FormatOrganize{
			Flags: organize_text.MakeFlags(),
		},
	)
}

type FormatOrganize struct {
	command_components.LocalWorkingCopy

	Flags organize_text.Flags
}

func (cmd *FormatOrganize) SetFlagSet(f *flag.FlagSet) {
	cmd.Flags.SetFlagSet(f)
}

func (cmd *FormatOrganize) Run(dep command.Request) {
	args := dep.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	cmd.Flags.Config = localWorkingCopy.GetConfig()

	if len(args) != 1 {
		localWorkingCopy.CancelWithErrorf("expected exactly one input argument")
	}

	var fdee fd.FD

	if err := fdee.Set(args[0]); err != nil {
		localWorkingCopy.CancelWithError(err)
	}

	var r io.Reader

	if fdee.IsStdin() {
		r = os.Stdin
	} else {
		var f *os.File

		{
			var err error

			if f, err = files.Open(args[0]); err != nil {
				localWorkingCopy.CancelWithError(err)
			}
		}

		r = f

		defer localWorkingCopy.MustClose(f)
	}

	var ot *organize_text.Text

	readOrganizeTextOp := user_ops.ReadOrganizeFile{}

	var repoId ids.RepoId

	{
		var err error

		if ot, err = readOrganizeTextOp.Run(
			localWorkingCopy,
			r,
			organize_text.NewMetadata(repoId),
		); err != nil {
			localWorkingCopy.CancelWithError(err)
		}
	}

	ot.Options = cmd.Flags.GetOptionsWithMetadata(
		localWorkingCopy.GetConfig().GetCLIConfig().PrintOptions,
		localWorkingCopy.SkuFormatBoxCheckedOutNoColor(),
		localWorkingCopy.GetStore().GetAbbrStore().GetAbbr(),
		sku.ObjectFactory{},
		ot.Metadata,
	)

	if err := ot.Refine(); err != nil {
		localWorkingCopy.CancelWithError(err)
	}

	if _, err := ot.WriteTo(os.Stdout); err != nil {
		localWorkingCopy.CancelWithError(err)
	}
}
