package string_format_writer

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/options_print"
)

func StringPrefixFromOptions(
	options options_print.Options,
) string {
	if options.Newlines {
		return "\n  " + StringIndent
	} else {
		return " "
	}
}

func WriteStringPrefixFormat(
	w interfaces.WriterAndStringWriter,
	prefix, body string,
) (n int64, err error) {
	var n1 int

	n1, err = w.WriteString(prefix)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = w.WriteString(body)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
