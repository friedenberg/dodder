package descriptions

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
)

type formatCli[T interfaces.Stringer] struct {
	*formatCliStringer
}

func MakeCliFormat(
	truncate string_format_writer.CliFormatTruncation,
	co string_format_writer.ColorOptions,
	quote bool,
) *formatCli[*Description] {
	return MakeCliFormatGeneric[*Description](
		truncate,
		co,
		quote,
	)
}

func MakeCliFormatGeneric[T interfaces.Stringer](
	truncate string_format_writer.CliFormatTruncation,
	co string_format_writer.ColorOptions,
	quote bool,
) *formatCli[T] {
	return &formatCli[T]{
		formatCliStringer: MakeCliFormatStringer(
			truncate,
			co,
			quote,
		),
	}
}

func (f *formatCli[T]) EncodeStringTo(
	k T,
	w interfaces.WriterAndStringWriter,
) (n int64, err error) {
	return f.formatCliStringer.EncodeStringTo(k, w)
}
