package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

func (local *Repo) ReadObjectHistory(
	objectId *ids.ObjectId,
) (objects []*sku.Transacted, err error) {
	streamIndex := local.GetStore().GetStreamIndex()

	if objects, err = streamIndex.ReadManyObjectId(
		objectId,
	); err != nil {
		err = errors.Wrap(err)
		return objects, err
	}

	return objects, err
}
