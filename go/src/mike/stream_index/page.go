package stream_index

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/golf/page_id"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type page struct {
	index *Index

	writeLock        sync.Mutex
	pageId           page_id.PageId
	forceFullWrite   bool
	additionsHistory pageAdditions
	additionsLatest  pageAdditions
}

func (page *page) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.index = index
	page.pageId = pageId
	page.additionsHistory.initialize(index)
	page.additionsLatest.initialize(index)
}

func (page *page) objectIdStringExists(objectIdString string) bool {
	if _, ok := page.additionsHistory.objectIdLookup[objectIdString]; ok {
		return true
	}

	if _, ok := page.additionsLatest.objectIdLookup[objectIdString]; ok {
		return true
	}

	return false
}

func (page *page) hasChanges() bool {
	return page.additionsHistory.hasChanges() ||
		page.additionsLatest.hasChanges() || page.forceFullWrite
}

func (page *page) lenAdded() int {
	return page.additionsHistory.Len() + page.additionsLatest.Len()
}

func (page *page) add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	additions := page.additionsHistory

	// TODO write binary representation to file-backed buffered writer and then
	// merge streams using raw binary data

	if page.index.sunrise.Less(object.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		additions = page.additionsLatest
	}

	additions.add(object)

	return err
}
