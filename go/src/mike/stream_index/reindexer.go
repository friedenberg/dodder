package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type Reindexer struct {
	index *Index
}

var _ sku.Reindexer = &Reindexer{}

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

func (reindexer *Reindexer) ReadOneMarklIdAdded(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (ok bool) {
	panic(errors.Err405MethodNotAllowed)
}

func (reindexer *Reindexer) ReadOneMarklId(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (ok bool) {
	return reindexer.index.ReadOneMarklId(marklId, object)
}
