package format

import (
	"fmt"
	"io"
	"strings"
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/string_format_writer"
)

func MakeFormatStringRightAligned(
	f string,
	args ...any,
) interfaces.FuncWriter {
	return func(w io.Writer) (n int64, err error) {
		f = fmt.Sprintf(f+" ", args...)

		diff := string_format_writer.LenStringMax + 1 - utf8.RuneCountInString(
			f,
		)

		if diff > 0 {
			f = strings.Repeat(" ", diff) + f
		}

		var n1 int

		if n1, err = io.WriteString(w, f); err != nil {
			n = int64(n1)
			err = errors.Wrap(err)
			return
		}

		n = int64(n1)

		return
	}
}
