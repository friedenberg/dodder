package inventory_list_blobs

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type V2IterDecoder struct {
	V2
}

func (coder V2IterDecoder) DecodeFrom(
	yield func(*sku.Transacted) bool,
	bufferedReader *bufio.Reader,
) (n int64, err error) {
	for {
		object := sku.GetTransactedPool().Get()
		// TODO Fix upstream issues with repooling
		// defer sku.GetTransactedPool().Put(object)

		if _, err = coder.V2ObjectCoder.DecodeFrom(object, bufferedReader); err != nil {
			if errors.IsEOF(err) {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return
			}
		}

		if !yield(object) {
			return
		}
	}

	return
}
