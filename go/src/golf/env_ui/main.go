package env_ui

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/debug"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/foxtrot/config_mutable_cli"
)

// TODO explore storing buffered writer and reader
type Env interface {
	// TODO remove and keep separate
	errors.Context

	GetOptions() Options
	GetIn() fd.Std
	GetInFile() io.Reader
	GetUI() fd.Std
	GetUIFile() interfaces.WriterAndStringWriter
	GetOut() fd.Std
	GetOutFile() interfaces.WriterAndStringWriter
	GetErr() fd.Std
	GetErrFile() interfaces.WriterAndStringWriter
	GetCLIConfig() config_mutable_cli.Config

	Confirm(message string) (success bool)
	Retry(header, retry string, err error) (tryAgain bool)

	FormatOutputOptions() (o string_format_writer.OutputOptions)
	FormatColorOptionsOut() (o string_format_writer.ColorOptions)
	FormatColorOptionsErr() (o string_format_writer.ColorOptions)
	StringFormatWriterFields(
		truncate string_format_writer.CliFormatTruncation,
		co string_format_writer.ColorOptions,
	) interfaces.StringEncoderTo[string_format_writer.Box]
}

type env struct {
	errors.Context

	options Options

	in  fd.Std
	ui  fd.Std
	out fd.Std
	err fd.Std

	debug *debug.Context

	cliConfig config_mutable_cli.Config
}

func MakeDefault() *env {
	return Make(
		errors.MakeContextDefault(),
		config_mutable_cli.Config{},
		Options{},
	)
}

func Make(
	context errors.Context,
	kCli config_mutable_cli.Config,
	options Options,
) *env {
	e := &env{
		Context:   context,
		options:   options,
		in:        fd.MakeStd(os.Stdin),
		out:       fd.MakeStd(os.Stdout),
		err:       fd.MakeStd(os.Stderr),
		cliConfig: kCli,
	}

	if options.UIFileIsStderr {
		e.ui = e.err
	} else {
		e.ui = e.out
	}

	{
		var err error

		if e.debug, err = debug.MakeContext(context, kCli.Debug); err != nil {
			context.CancelWithError(err)
		}
	}

	if kCli.Verbose && !kCli.Quiet {
		ui.SetVerbose(true)
	} else {
		ui.SetOutput(io.Discard)
	}

	if kCli.Todo {
		ui.SetTodoOn()
	}

	return e
}

func (u env) GetOptions() Options {
	return u.options
}

func (u *env) GetIn() fd.Std {
	return u.in
}

func (u *env) GetInFile() io.Reader {
	return u.in.GetFile()
}

func (u *env) GetUI() fd.Std {
	return u.ui
}

func (u *env) GetUIFile() interfaces.WriterAndStringWriter {
	return u.ui.GetFile()
}

func (u *env) GetOut() fd.Std {
	return u.out
}

func (u *env) GetOutFile() interfaces.WriterAndStringWriter {
	return u.out.GetFile()
}

func (u *env) GetErr() fd.Std {
	return u.err
}

func (u *env) GetErrFile() interfaces.WriterAndStringWriter {
	return u.err.GetFile()
}

func (u *env) GetCLIConfig() config_mutable_cli.Config {
	return u.cliConfig
}
