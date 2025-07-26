package inventory_list_blobs

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func ReadInventoryListBlob(
	listFormat sku.ListFormat,
	reader *bufio.Reader,
	list *sku.List,
) (err error) {
	iter := listFormat.StreamInventoryListBlobSkus(reader)
	
	for sk, iterErr := range iter {
		if iterErr != nil {
			err = errors.Wrap(iterErr)
			return
		}

		if err = sk.CalculateObjectShas(); err != nil {
			err = errors.Wrap(err)
			return
		}

		if err = list.Add(sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
