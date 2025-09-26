package stream_index

import (
	"sync"

	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
)

// TODO rename to ???
type page struct {
	writeLock sync.Mutex
	pageId    page_id.PageId
	additions pageAdditions
}

func (page *page) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.pageId = pageId
	page.additions.initialize()
}
