package inventory_list_coders

import (
	"bufio"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type doddishV1 struct {
	Box *box_format.BoxTransacted
}

func (coder doddishV1) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if object.Metadata.GetObjectDigest().IsNull() {
		err = errors.ErrorWithStackf("empty sha: %q", sku.String(object))
		return
	}

	var n1 int64
	var n2 int

	n1, err = coder.Box.EncodeStringTo(object, bufferedWriter)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = fmt.Fprintf(bufferedWriter, "\n")
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (coder doddishV1) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var isEOF bool

	if n, err = coder.Box.ReadStringFormat(object, bufferedReader); err != nil {
		if err == io.EOF {
			isEOF = true

			if n == 0 {
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	if err = object.CalculateObjectDigests(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if isEOF {
		err = io.EOF
	}

	return
}
