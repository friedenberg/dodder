package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/checked_out_state"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/foxtrot/id_fmts"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

// TODO migrate to StringFormatWriterSkuBoxCheckedOut
func (repo *Repo) PrinterTransactedDeleted() interfaces.FuncIter[*sku.CheckedOut] {
	printOptions := repo.config.GetCLIConfig().PrintOptions.
		WithPrintShas(true).
		WithPrintTime(false)

	stringEncoder := repo.StringFormatWriterSkuBoxCheckedOut(
		printOptions,
		repo.FormatColorOptionsOut(),
		string_format_writer.CliFormatTruncation66CharEllipsis,
		box_format.CheckedOutHeaderDeleted{
			ConfigDryRunReader: repo.GetConfig().GetCLIConfig(),
		},
	)

	return string_format_writer.MakeDelim(
		"\n",
		repo.GetUIFile(),
		string_format_writer.MakeFunc(
			func(
				writer interfaces.WriterAndStringWriter,
				object *sku.CheckedOut,
			) (n int64, err error) {
				return stringEncoder.EncodeStringTo(object, writer)
			},
		),
	)
}

// TODO make generic external version
func (repo *Repo) PrinterFDDeleted() interfaces.FuncIter[*fd.FD] {
	p := id_fmts.MakeFDDeletedStringWriterFormat(
		repo.GetConfig().GetCLIConfig().IsDryRun(),
		id_fmts.MakeFDCliFormat(
			repo.FormatColorOptionsOut(),
			repo.envRepo.MakeRelativePathStringFormatWriter(),
		),
	)

	return string_format_writer.MakeDelim(
		"\n",
		repo.GetUIFile(),
		p,
	)
}

func (u *Repo) PrinterHeader() interfaces.FuncIter[string] {
	if u.config.GetCLIConfig().PrintOptions.PrintFlush {
		return string_format_writer.MakeDelim(
			"\n",
			u.GetErrFile(),
			string_format_writer.MakeDefaultDatePrefixFormatWriter(
				u,
				string_format_writer.MakeColor(
					u.FormatColorOptionsOut(),
					string_format_writer.MakeString[string](),
					string_format_writer.ColorTypeHeading,
				),
			),
		)
	} else {
		return func(v string) error { return ui.Log().Print(v) }
	}
}

func (u *Repo) PrinterCheckedOutConflictsForRemoteTransfers() interfaces.FuncIter[*sku.CheckedOut] {
	p := u.PrinterCheckedOut(box_format.CheckedOutHeaderState{})

	return func(co *sku.CheckedOut) (err error) {
		if co.GetState() != checked_out_state.Conflicted {
			return
		}

		if err = p(co); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}
}

func (u *Repo) MakePrinterBoxArchive(
	out interfaces.WriterAndStringWriter,
	includeTai bool,
) interfaces.FuncIter[*sku.Transacted] {
	boxFormat := box_format.MakeBoxTransactedArchive(
		u.GetEnv(),
		u.GetConfig().GetCLIConfig().PrintOptions.WithPrintTai(includeTai),
	)

	return string_format_writer.MakeDelim(
		"\n",
		out,
		string_format_writer.MakeFunc(
			func(w interfaces.WriterAndStringWriter, o *sku.Transacted) (n int64, err error) {
				return boxFormat.EncodeStringTo(o, w)
			},
		),
	)
}
