package stream_index

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO rename to ???
type writtenPage struct {
	writeLock sync.Mutex
	pageId    page_id.PageId
	additions pageAdditions
}

type pageAdditions struct {
	forceFullFlush      bool
	addedObjectIdLookup map[string]struct{}
	added, addedLatest  *sku.ListTransacted
}

func (page *writtenPage) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.pageId = pageId
	page.additions.initialize()
}

func (additions *pageAdditions) initialize() {
	additions.added = sku.MakeListTransacted()
	additions.addedLatest = sku.MakeListTransacted()
	additions.addedObjectIdLookup = make(map[string]struct{})
}

// TODO write binary representation to file-backed buffered writer and then
// merge streams using raw binary data
func (index *Index) add(
	pageIndex PageIndex,
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	pageAdditions := &index.pages[pageIndex].additions

	pageAdditions.addedObjectIdLookup[object.ObjectId.String()] = struct{}{}
	objectClone := object.CloneTransacted()

	if index.sunrise.Less(objectClone.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		pageAdditions.addedLatest.Add(objectClone)
	} else {
		pageAdditions.added.Add(objectClone)
	}

	return err
}

func (additions *pageAdditions) hasChanges() bool {
	return additions.waitingToAddLen() > 0 || additions.forceFullFlush
}

func (additions *pageAdditions) waitingToAddLen() int {
	return additions.added.Len() + additions.addedLatest.Len()
}
