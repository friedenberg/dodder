package store

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/store_browser"
	"code.linenisgreat.com/dodder/go/src/lima/store_fs"
)

// TODO remove entirely
func (s *Store) PutCheckedOutLike(co sku.SkuType) {
	switch co.GetSkuExternal().GetRepoId().GetRepoIdString() {
	// TODO make generic?
	case "browser":
		store_browser.GetCheckedOutPool().Put(co)

	default:
		store_fs.GetCheckedOutPool().Put(co)
	}
}
