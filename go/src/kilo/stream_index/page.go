package stream_index

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
)

type page struct {
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
	page.pageId = pageId
	page.additionsHistory.initialize()
	page.additionsLatest.initialize()
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
