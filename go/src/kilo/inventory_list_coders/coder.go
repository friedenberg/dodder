package inventory_list_coders

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type coder struct {
	sku.ListCoder
	beforeEncoding func(*sku.Transacted) error
	afterDecoding  func(*sku.Transacted) error
}

func (coder coder) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if coder.beforeEncoding != nil {
		if err = coder.beforeEncoding(object); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if n, err = coder.ListCoder.EncodeTo(object, bufferedWriter); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (coder coder) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	if n, err = coder.ListCoder.DecodeFrom(object, bufferedReader); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	if coder.afterDecoding != nil {
		if err = coder.afterDecoding(object); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
