package commands

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/checkout_mode"
	"code.linenisgreat.com/dodder/go/src/charlie/checkout_options"
	"code.linenisgreat.com/dodder/go/src/charlie/script_config"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/object_metadata"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"format-object",
		&FormatObject{
			CheckoutMode: checkout_mode.BlobOnly,
		},
	)
}

type FormatObject struct {
	command_components.LocalWorkingCopy

	CheckoutMode checkout_mode.Mode // add test that says this is unused for stdin
	Stdin        bool               // switch to using `-`
	ids.RepoId
	UTIGroup string
}

var _ interfaces.CommandComponentWriter = (*FormatObject)(nil)

func (cmd *FormatObject) SetFlagDefinitions(f interfaces.CommandLineFlagDefinitions) {
	f.BoolVar(
		&cmd.Stdin,
		"stdin",
		false,
		"Read object from stdin and use a Type directly",
	)

	f.Var(&cmd.RepoId, "kasten", "none or Browser")

	f.StringVar(&cmd.UTIGroup, "uti-group", "", "lookup format from UTI group")

	f.Var(&cmd.CheckoutMode, "mode", "mode for checking out the zettel")
}

func (cmd *FormatObject) Run(req command.Request) {
	args := req.PopArgs()
	localWorkingCopy := cmd.MakeLocalWorkingCopy(req)

	if cmd.Stdin {
		if err := cmd.FormatFromStdin(localWorkingCopy, args...); err != nil {
			localWorkingCopy.Cancel(err)
		}

		return
	}

	var formatId string

	var objectIdString string
	var blobFormatter script_config.RemoteScript

	switch len(args) {
	case 2:
		formatId = args[1]
		fallthrough

	case 1:
		objectIdString = args[0]

	default:
		errors.ContextCancelWithErrorf(
			localWorkingCopy,
			"expected one or two input arguments, but got %d",
			len(args),
		)
	}

	var object *sku.Transacted

	{
		var err error

		if object, err = localWorkingCopy.GetZettelFromObjectId(
			objectIdString,
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	tipe := object.GetType()

	{
		var err error

		if blobFormatter, err = localWorkingCopy.GetBlobFormatter(
			tipe,
			formatId,
			cmd.UTIGroup,
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	formatter := typed_blob_store.MakeTextFormatterWithBlobFormatter(
		localWorkingCopy.GetEnvRepo(),
		checkout_options.TextFormatterOptions{
			DoNotWriteEmptyDescription: true,
		},
		localWorkingCopy.GetConfig(),
		blobFormatter,
		checkout_mode.None,
	)

	if err := localWorkingCopy.GetStore().TryFormatHook(object); err != nil {
		localWorkingCopy.Cancel(err)
	}

	if _, err := formatter.WriteStringFormatWithMode(
		localWorkingCopy.GetUIFile(),
		object,
		cmd.CheckoutMode,
	); err != nil {
		var errBlobFormatterFailed *object_metadata.ErrBlobFormatterFailed

		if errors.As(err, &errBlobFormatterFailed) {
			localWorkingCopy.Cancel(errBlobFormatterFailed)
			// err = nil
			// ui.Err().Print(errExit)
		} else {
			localWorkingCopy.Cancel(err)
		}
	}
}

func (c *FormatObject) FormatFromStdin(
	u *local_working_copy.Repo,
	args ...string,
) (err error) {
	formatId := "text"

	var blobFormatter script_config.RemoteScript
	var tipe ids.Type

	switch len(args) {
	case 1:
		if err = tipe.Set(args[0]); err != nil {
			err = errors.Wrap(err)
			return err
		}

	case 2:
		formatId = args[0]
		if err = tipe.Set(args[1]); err != nil {
			err = errors.Wrap(err)
			return err
		}

	default:
		err = errors.ErrorWithStackf(
			"expected one or two input arguments, but got %d",
			len(args),
		)
		return err
	}

	if blobFormatter, err = u.GetBlobFormatter(
		tipe,
		formatId,
		c.UTIGroup,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	var wt io.WriterTo

	if wt, err = script_config.MakeWriterToWithStdin(
		blobFormatter,
		u.GetEnvRepo().MakeCommonEnv(),
		u.GetInFile(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if _, err = wt.WriteTo(u.GetUIFile()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
