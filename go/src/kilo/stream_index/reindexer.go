package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Reindexer struct {
	index *Index
}

var _ IndexCommon = &Reindexer{}

func (reindexer *Reindexer) Add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	return reindexer.index.Add(object, options)
}

func (reindexer *Reindexer) ObjectExists(
	objectId *ids.ObjectId,
) (err error) {
	return reindexer.index.ObjectExists(objectId)
}

func (reindexer *Reindexer) ReadOneObjectId(
	objectId interfaces.ObjectId,
	object *sku.Transacted,
) (err error) {
	return reindexer.index.ReadOneObjectId(objectId, object)
}
