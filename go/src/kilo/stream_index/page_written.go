package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type writtenPage struct {
	pageId     page_id.PageId
	hasChanges bool

	// TODO separate
	addedObjectIdLookup map[string]struct{}
	added, addedLatest  *sku.ListTransacted
}

func (page *writtenPage) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.pageId = pageId
	page.added = sku.MakeListTransacted()
	page.addedLatest = sku.MakeListTransacted()
	page.addedObjectIdLookup = make(map[string]struct{})
}

// TODO write binary representation to file-backed buffered writer and then
// merge streams using raw binary data
func (index *Index) add(
	pageIndex PageIndex,
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	page := &index.pages[pageIndex]

	page.addedObjectIdLookup[object.ObjectId.String()] = struct{}{}
	objectClone := object.CloneTransacted()

	if index.sunrise.Less(objectClone.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		page.addedLatest.Add(objectClone)
	} else {
		page.added.Add(objectClone)
	}

	page.hasChanges = true

	return err
}

func (page *writtenPage) waitingToAddLen() int {
	return page.added.Len() + page.addedLatest.Len()
}
