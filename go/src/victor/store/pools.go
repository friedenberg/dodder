package store

import (
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/romeo/store_fs"
	"code.linenisgreat.com/dodder/go/src/uniform/store_browser"
)

// TODO remove entirely
func (store *Store) PutCheckedOutLike(co sku.SkuType) {
	switch co.GetSkuExternal().GetRepoId().GetRepoIdString() {
	// TODO make generic?
	case "browser":
		store_browser.GetCheckedOutPool().Put(co)

	default:
		store_fs.GetCheckedOutPool().Put(co)
	}
}
