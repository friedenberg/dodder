package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type writtenPage struct {
	pageId    page_id.PageId
	additions pageAdditions
}

type pageAdditions struct {
	hasChanges          bool
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

func (pageAdditions *pageAdditions) initialize() {
	pageAdditions.added = sku.MakeListTransacted()
	pageAdditions.addedLatest = sku.MakeListTransacted()
	pageAdditions.addedObjectIdLookup = make(map[string]struct{})
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

	pageAdditions.hasChanges = true

	return err
}

func (pageAdditions *pageAdditions) getHasChanges() bool {
	return pageAdditions.hasChanges
}

func (pageAdditions *pageAdditions) waitingToAddLen() int {
	return pageAdditions.added.Len() + pageAdditions.addedLatest.Len()
}
