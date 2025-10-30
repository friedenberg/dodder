package env_ui

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/foxtrot/repo_config_cli"
)

// TODO explore storing buffered writer and reader
type Env interface {
	// TODO remove and keep separate
	interfaces.Context

	GetOptions() Options
	GetIn() fd.Std
	GetInFile() io.Reader
	GetUI() fd.Std
	GetUIFile() interfaces.WriterAndStringWriter
	GetOut() fd.Std
	GetOutFile() interfaces.WriterAndStringWriter
	GetErr() fd.Std
	GetErrFile() interfaces.WriterAndStringWriter
	GetCLIConfig() repo_config_cli.Config

	Confirm(message string) (success bool)
	Retry(header, retry string, err error) (tryAgain bool)

	FormatOutputOptions(
		options_print.Options,
	) (o string_format_writer.OutputOptions)
	FormatColorOptionsOut(
		options_print.Options,
	) (o string_format_writer.ColorOptions)
	FormatColorOptionsErr(
		options_print.Options,
	) (o string_format_writer.ColorOptions)
	StringFormatWriterFields(
		truncate string_format_writer.CliFormatTruncation,
		co string_format_writer.ColorOptions,
	) interfaces.StringEncoderTo[string_format_writer.Box]
}

type env struct {
	interfaces.Context

	options Options

	in  fd.Std
	ui  fd.Std
	out fd.Std
	err fd.Std

	debug *debug.Context

	cliConfig repo_config_cli.Config
}

func MakeDefault(ctx interfaces.Context) *env {
	return Make(
		ctx,
		repo_config_cli.Config{},
		Options{},
	)
}

func Make(
	context interfaces.Context,
	cliConfig repo_config_cli.Config,
	options Options,
) *env {
	// TODO use ui printing prefix
	env := &env{
		Context:   context,
		options:   options,
		in:        fd.MakeStd(os.Stdin),
		out:       fd.MakeStd(os.Stdout),
		err:       fd.MakeStd(os.Stderr),
		cliConfig: cliConfig,
	}

	if options.UIFileIsStderr {
		env.ui = env.err
	} else {
		env.ui = env.out
	}

	{
		var err error

		if env.debug, err = debug.MakeContext(context, cliConfig.Debug); err != nil {
			context.Cancel(err)
		}
	}

	if cliConfig.Verbose && !cliConfig.Quiet {
		ui.SetVerbose(true)
	} else {
		ui.SetOutput(io.Discard)
	}

	if cliConfig.Todo {
		ui.SetTodoOn()
	}

	return env
}

func (env env) GetOptions() Options {
	return env.options
}

func (env *env) GetIn() fd.Std {
	return env.in
}

func (env *env) GetInFile() io.Reader {
	return env.in.GetFile()
}

func (env *env) GetUI() fd.Std {
	return env.ui
}

func (env *env) GetUIFile() interfaces.WriterAndStringWriter {
	return env.ui.GetFile()
}

func (env *env) GetOut() fd.Std {
	return env.out
}

func (env *env) GetOutFile() interfaces.WriterAndStringWriter {
	return env.out.GetFile()
}

func (env *env) GetErr() fd.Std {
	return env.err
}

func (env *env) GetErrFile() interfaces.WriterAndStringWriter {
	return env.err.GetFile()
}

func (env *env) GetCLIConfig() repo_config_cli.Config {
	return env.cliConfig
}
