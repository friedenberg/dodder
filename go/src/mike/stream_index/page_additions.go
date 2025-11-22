package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
)

type pageAdditions struct {
	objectIdLookup map[string]struct{}
	objects        *sku.HeapTransactedTai
	// objects        *sku.OpenList
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

func (additions *pageAdditions) Reset() {
	additions.objects.Reset()
}

func (additions *pageAdditions) All() interfaces.Seq[*sku.Transacted] {
	return additions.objects.All()
}
