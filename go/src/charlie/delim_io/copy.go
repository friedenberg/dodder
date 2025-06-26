package delim_io

import (
	"fmt"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
)

// Copies each `delim` suffixed segment from src to dst, and for each segment,
// adds the passed in prefix string.
//
// Useful for taking a Reader and adding a prefix for every line, like how `git`
// shows `remote: <line>` for all remote stderr output.
// TODO extract into an io.Writer-like object
func CopyWithPrefixOnDelim(
	delimiter byte,
	prefix string,
	destination ui.Printer,
	source io.Reader,
	includeLineNumbers bool,
) (n int64, err error) {
	bufferedReader := pool.GetBufioReader().Get()
	defer pool.GetBufioReader().Put(bufferedReader)

	bufferedReader.Reset(source)

	var (
		isEOF    bool
		lineNumber int
	)

	var stringBuilder strings.Builder

	for !isEOF {
		var rawLine string

		rawLine, err = bufferedReader.ReadString(delimiter)
		n1 := len(rawLine)
		n += int64(n1)

		if err != nil && !errors.IsEOF(err) {
			err = errors.Wrap(err)
			return
		}

		if errors.IsEOF(err) {
			isEOF = true
			err = nil

			if n1 == 0 {
				break
			}
		}

		stringBuilder.WriteString(prefix)
		fmt.Fprint(&stringBuilder, ":")

		if includeLineNumbers {
			fmt.Fprintf(&stringBuilder, "%d:", lineNumber)
		}

		fmt.Fprint(&stringBuilder, " ")
		// fmt.Fprint(bw, "\t")

		stringBuilder.WriteString(
			strings.TrimSuffix(rawLine, string([]byte{delimiter})),
		)

		destination.Print(stringBuilder.String())
		stringBuilder.Reset()

		lineNumber++
	}

	return
}
