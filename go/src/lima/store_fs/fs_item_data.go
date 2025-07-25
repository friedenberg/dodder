package store_fs

import (
	"maps"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type fsItemData struct {
	interfaces.MutableSetLike[*sku.FSItem]
	shas map[string]interfaces.MutableSetLike[*sku.FSItem]
}

func makeFSItemData() fsItemData {
	return fsItemData{
		MutableSetLike: collections_value.MakeMutableValueSet[*sku.FSItem](nil),
		shas:           make(map[string]interfaces.MutableSetLike[*sku.FSItem]),
	}
}

func (src *fsItemData) Clone() (dst fsItemData) {
	dst.MutableSetLike = src.MutableSetLike.CloneMutableSetLike()
	dst.shas = maps.Clone(src.shas)
	return
}

func (data *fsItemData) ConsolidateDuplicateBlobs() (err error) {
	replacement := collections_value.MakeMutableValueSet[*sku.FSItem](nil)

	for _, fds := range data.shas {
		if fds.Len() == 1 {
			replacement.Add(fds.Any())
		}

		sorted := quiter.ElementsSorted(
			fds,
			func(a, b *sku.FSItem) bool {
				return a.ExternalObjectId.String() < b.ExternalObjectId.String()
			},
		)

		top := sorted[0]

		for _, other := range sorted[1:] {
			for item := range other.MutableSetLike.All() {
				top.MutableSetLike.Add(item)
			}
		}

		replacement.Add(top)
	}

	// TODO make less leaky
	data.MutableSetLike = replacement

	return
}
