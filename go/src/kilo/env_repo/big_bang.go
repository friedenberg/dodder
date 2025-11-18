package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/india/genesis_configs"
)

// Config used to initialize a repo for the first time
type BigBang struct {
	GenesisConfig        *genesis_configs.TypedConfigPrivateMutable
	TypedBlobStoreConfig *blob_store_configs.TypedMutableConfig

	InventoryListType ids.Type

	Yin                  string
	Yang                 string
	ExcludeDefaultType   bool
	ExcludeDefaultConfig bool
	OverrideXDGWithCwd   bool
	BlobStoreId          blob_store_id.Id
}

func (bigBang *BigBang) SetDefaults() {
	bigBang.GenesisConfig = genesis_configs.Default()
	bigBang.InventoryListType = ids.GetOrPanic(
		ids.TypeInventoryListVCurrent,
	).Type

	bigBang.TypedBlobStoreConfig = blob_store_configs.Default()
}
