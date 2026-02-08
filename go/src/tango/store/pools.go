package store

import (
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/papa/store_fs"
	"code.linenisgreat.com/dodder/go/src/sierra/store_browser"
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
