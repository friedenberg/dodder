package commands

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/organize_text"
	"code.linenisgreat.com/dodder/go/src/papa/command_components_dodder"
	"code.linenisgreat.com/dodder/go/src/papa/user_ops"
)

func init() {
	utility.AddCmd(
		"format-organize",
		&FormatOrganize{
			Flags: organize_text.MakeFlags(),
		})
}

type FormatOrganize struct {
	command_components_dodder.LocalWorkingCopy

	Flags organize_text.Flags
}

var _ interfaces.CommandComponentWriter = (*FormatOrganize)(nil)

func (cmd *FormatOrganize) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
	cmd.Flags.SetFlagDefinitions(f)
}

func (cmd *FormatOrganize) Run(dep command.Request) {
	args := dep.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(dep)

	cmd.Flags.Config = localWorkingCopy.GetConfigPtr()

	if len(args) != 1 {
		errors.ContextCancelWithErrorf(
			localWorkingCopy,
			"expected exactly one input argument",
		)
	}

	var fdee fd.FD

	if err := fdee.Set(args[0]); err != nil {
		localWorkingCopy.Cancel(err)
	}

	var r io.Reader

	if fdee.IsStdin() {
		r = os.Stdin
	} else {
		var f *os.File

		{
			var err error

			if f, err = files.Open(args[0]); err != nil {
				localWorkingCopy.Cancel(err)
			}
		}

		r = f

		defer errors.ContextMustClose(localWorkingCopy, f)
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
			localWorkingCopy.Cancel(err)
		}
	}

	ot.Options = cmd.Flags.GetOptionsWithMetadata(
		localWorkingCopy.GetConfig().GetPrintOptions(),
		localWorkingCopy.SkuFormatBoxCheckedOutNoColor(),
		localWorkingCopy.GetStore().GetAbbrStore().GetAbbr(),
		sku.ObjectFactory{},
		ot.Metadata,
	)

	if err := ot.Refine(); err != nil {
		localWorkingCopy.Cancel(err)
	}

	if _, err := ot.WriteTo(os.Stdout); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
