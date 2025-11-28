package stream_index

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type pageAdditions struct {
	defaultObjectDigestMarklFormatId string
	index                            *Index
	objectIdLookup                   map[string]struct{}
	objects                          *sku.HeapTransacted
	// objects        *sku.OpenList
}

func (additions *pageAdditions) initialize(index *Index) {
	index.defaultObjectDigestMarklFormatId = index.envRepo.GetObjectDigestType()

	additions.index = index
	additions.objects = sku.MakeListTransacted()
	additions.objectIdLookup = make(map[string]struct{})
}

func (additions *pageAdditions) add(object *sku.Transacted) {
	objectClone := object.CloneTransacted()

	additions.objects.Add(objectClone)
	additions.objectIdLookup[object.ObjectId.String()] = struct{}{}

	seqProbeIds := object.AllProbeIds(
		additions.index.index.GetHashType(),
		additions.defaultObjectDigestMarklFormatId,
	)

	additionProbes := additions.index.probeIndex.additionProbes

	for probeId := range seqProbeIds {
		idBytes := probeId.Id.GetBytes()
		additionProbes.Set(string(idBytes), objectClone)
	}
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
