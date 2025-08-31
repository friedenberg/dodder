package inventory_list_coders

import (
	"bufio"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
)

type doddishV2 struct {
	box           *box_format.BoxTransacted
	genesisConfig genesis_configs.ConfigPrivate
}

func (coder doddishV2) EncodeTo(
	object *sku.Transacted,
	bufferedWriter *bufio.Writer,
) (n int64, err error) {
	if err = markl.AssertIdIsNotNull(
		object.Metadata.GetObjectDigest(),
		"object-digest",
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", sku.String(object))
		return
	}

	if err = markl.AssertIdIsNotNull(
		object.Metadata.GetObjectSig(),
		"object-sig",
	); err != nil {
		err = errors.Wrapf(err, "Object: %q", sku.String(object))
		return
	}

	var n1 int64
	var n2 int

	n1, err = coder.box.EncodeStringTo(object, bufferedWriter)
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

func (coder doddishV2) DecodeFrom(
	object *sku.Transacted,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	var isEOF bool

	if n, err = coder.box.ReadStringFormat(object, bufferedReader); err != nil {
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

	if err = object.FinalizeAndVerify(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if isEOF {
		err = io.EOF
	}

	return
}
