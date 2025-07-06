package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func (local *Repo) ReadObjectHistory(
	oid *ids.ObjectId,
) (skus []*sku.Transacted, err error) {
	streamIndex := local.GetStore().GetStreamIndex()

	if skus, err = streamIndex.ReadManyObjectId(
		oid,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
