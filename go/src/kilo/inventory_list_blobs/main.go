package inventory_list_blobs

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

var (
	_ sku.ListFormat = V1{}
	_ sku.ListFormat = V2{}
)

func WriteObjectToOpenList(
	format sku.ListFormat,
	object *sku.Transacted,
	list *sku.OpenList,
) (n int64, err error) {
	if !list.LastTai.Less(object.GetTai()) {
		err = errors.Errorf(
			"object order incorrect. Last: %s, current: %s",
			list.LastTai,
			object.GetTai(),
		)

		return
	}

	bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(list.Mover)
	defer repoolBufferedWriter()

	if n, err = format.EncodeTo(
		object,
		bufferedWriter,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	list.LastTai = object.GetTai()
	list.Len += 1

	return
}

// TODO also return a repool func
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
