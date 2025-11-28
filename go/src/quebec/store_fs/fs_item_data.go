package store_fs

import (
	"maps"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_value"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
)

type fsItemData struct {
	interfaces.MutableSetLike[*sku.FSItem]
	digests map[string]interfaces.MutableSetLike[*sku.FSItem]
}

func makeFSItemData() fsItemData {
	return fsItemData{
		MutableSetLike: collections_value.MakeMutableValueSet[*sku.FSItem](nil),
		digests:        make(map[string]interfaces.MutableSetLike[*sku.FSItem]),
	}
}

func (data *fsItemData) Clone() (dst fsItemData) {
	dst.MutableSetLike = data.MutableSetLike.CloneMutableSetLike()
	dst.digests = maps.Clone(data.digests)
	return dst
}

func (data *fsItemData) ConsolidateDuplicateBlobs() (err error) {
	replacement := collections_value.MakeMutableValueSet[*sku.FSItem](nil)

	for _, fds := range data.digests {
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
			for item := range other.FDs.All() {
				top.FDs.Add(item)
			}
		}

		replacement.Add(top)
	}

	// TODO make less leaky
	data.MutableSetLike = replacement

	return err
}
