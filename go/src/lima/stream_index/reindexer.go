package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type Reindexer struct {
	index *Index
	pages [PageCount]*pageAdditionsFileBacked
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
	return reindexer.index.ReadOneMarklIdAdded(marklId, object)
}

func (reindexer *Reindexer) ReadOneMarklId(
	marklId interfaces.MarklId,
	object *sku.Transacted,
) (ok bool) {
	return reindexer.index.ReadOneMarklId(marklId, object)
}
