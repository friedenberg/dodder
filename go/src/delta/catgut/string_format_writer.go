package catgut

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type stringFormatWriter struct{}

var StringFormatWriterString stringFormatWriter

func (stringFormatWriter) EncodeStringTo(
	e *String,
	sw interfaces.WriterAndStringWriter,
) (n int64, err error) {
	n, err = e.WriteTo(sw)
	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	return n, err
}
