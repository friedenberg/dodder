package env_ui

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
)

func (env *env) FormatOutputOptions() (o string_format_writer.OutputOptions) {
	o.ColorOptionsOut = env.FormatColorOptionsOut()
	o.ColorOptionsErr = env.FormatColorOptionsErr()
	return
}

func (env *env) shouldUseColorOutput(fd fd.Std) bool {
	if env.options.IgnoreTtyState {
		return env.cliConfig.PrintOptions.PrintColors
	} else {
		return fd.IsTty() && env.cliConfig.PrintOptions.PrintColors
	}
}

func (env *env) FormatColorOptionsOut() (o string_format_writer.ColorOptions) {
	o.OffEntirely = !env.shouldUseColorOutput(env.GetOut())
	return
}

func (env *env) FormatColorOptionsErr() (o string_format_writer.ColorOptions) {
	o.OffEntirely = !env.shouldUseColorOutput(env.GetErr())
	return
}

func (env *env) StringFormatWriterFields(
	truncate string_format_writer.CliFormatTruncation,
	co string_format_writer.ColorOptions,
) interfaces.StringEncoderTo[string_format_writer.Box] {
	return string_format_writer.MakeBoxStringEncoder(truncate, co)
}
