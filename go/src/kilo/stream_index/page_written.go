package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

type writtenPage struct {
	pageId              page_id.PageId
	sunrise             ids.Tai
	probeIndex          *probeIndex
	hasChanges          bool
	config              store_config.Store
	addedObjectIdLookup map[string]struct{}

	added, addedLatest *sku.ListTransacted
}

func (page *writtenPage) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.sunrise = index.sunrise
	page.pageId = pageId
	page.added = sku.MakeListTransacted()
	page.addedLatest = sku.MakeListTransacted()
	page.probeIndex = &index.probeIndex
	page.addedObjectIdLookup = make(map[string]struct{})
}

// TODO write binary representation to file-backed buffered writer and then
// merge streams using raw binary data
func (page *writtenPage) add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	page.addedObjectIdLookup[object.ObjectId.String()] = struct{}{}
	objectClone := object.CloneTransacted()

	if page.sunrise.Less(objectClone.GetTai()) ||
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
