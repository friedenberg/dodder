package env_ui

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

func (u *env) FormatOutputOptions() (o string_format_writer.OutputOptions) {
	o.ColorOptionsOut = u.FormatColorOptionsOut()
	o.ColorOptionsErr = u.FormatColorOptionsErr()
	return
}

func (u *env) shouldUseColorOutput(fd fd.Std) bool {
	if u.options.IgnoreTtyState {
		return u.cliConfig.PrintOptions.PrintColors
	} else {
		return fd.IsTty() && u.cliConfig.PrintOptions.PrintColors
	}
}

func (u *env) FormatColorOptionsOut() (o string_format_writer.ColorOptions) {
	o.OffEntirely = !u.shouldUseColorOutput(u.GetOut())
	return
}

func (u *env) FormatColorOptionsErr() (o string_format_writer.ColorOptions) {
	o.OffEntirely = !u.shouldUseColorOutput(u.GetErr())
	return
}

func (u *env) StringFormatWriterFields(
	truncate string_format_writer.CliFormatTruncation,
	co string_format_writer.ColorOptions,
) interfaces.StringEncoderTo[string_format_writer.Box] {
	return string_format_writer.MakeBoxStringEncoder(truncate, co)
}
