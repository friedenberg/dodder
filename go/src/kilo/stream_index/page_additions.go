package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

// TODO write binary representation to file-backed buffered writer and then
// merge streams using raw binary data
func (index *Index) add(
	pageIndex PageIndex,
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	page := &index.pages[pageIndex]

	additions := page.additionsHistory

	if index.sunrise.Less(object.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		additions = page.additionsLatest
	}

	additions.add(object)

	return err
}

type pageAdditions struct {
	objectIdLookup map[string]struct{}
	objects        *sku.ListTransacted
}

func (additions *pageAdditions) initialize() {
	additions.objects = sku.MakeListTransacted()
	additions.objectIdLookup = make(map[string]struct{})
}

func (additions *pageAdditions) add(object *sku.Transacted) {
	additions.objects.Add(object.CloneTransacted())
	additions.objectIdLookup[object.ObjectId.String()] = struct{}{}
}

func (additions *pageAdditions) hasChanges() bool {
	return additions.Len() > 0
}

func (additions *pageAdditions) Len() int {
	return additions.objects.Len()
}
