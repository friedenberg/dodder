package local_working_copy

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/dormant_index"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
	"code.linenisgreat.com/dodder/go/src/lima/env_lua"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/mike/store"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

func (local *Repo) GetEnv() env_ui.Env {
	return local
}

func (local *Repo) GetImmutableConfigPublic() genesis_configs.BlobPublic {
	return local.GetEnvRepo().GetConfigPublic().Blob
}

func (local *Repo) GetImmutableConfigPublicType() ids.Type {
	return local.GetEnvRepo().GetConfigPublic().Type
}

func (local *Repo) GetImmutableConfigPrivate() genesis_configs.TypedBlobPrivate {
	return local.GetEnvRepo().GetConfigPrivate()
}

func (local *Repo) GetEnvLocal() env_local.Env {
	return local
}

func (local *Repo) GetEnvWorkspace() env_workspace.Env {
	return local.envWorkspace
}

func (local *Repo) GetEnvLua() env_lua.Env {
	return local.envLua
}

func (local *Repo) GetTime() time.Time {
	return time.Now()
}

func (local *Repo) GetConfigStore() store_config.Store {
	return local.config
}

func (local *Repo) GetConfigStoreMutable() store_config.StoreMutable {
	return local.config
}

func (local *Repo) GetConfig() store_config.Config {
	return local.config.GetConfig()
}

func (local *Repo) GetConfigPtr() *store_config.Config {
	return local.config.GetConfigPtr()
}

func (local *Repo) GetDormantIndex() *dormant_index.Index {
	return &local.dormantIndex
}

func (local *Repo) GetEnvRepo() env_repo.Env {
	return local.envRepo
}

func (local *Repo) GetTypedInventoryListBlobStore() typed_blob_store.InventoryList {
	return local.typedBlobStore.InventoryList
}

func (local *Repo) GetBlobStore() interfaces.BlobStore {
	return local.GetEnvRepo().GetDefaultBlobStore()
}

func (local *Repo) GetObjectStore() sku.RepoStore {
	return &local.store
}

func (local *Repo) GetInventoryListStore() sku.InventoryListStore {
	return local.GetStore().GetInventoryListStore()
}

func (local *Repo) GetStore() *store.Store {
	return &local.store
}

func (local *Repo) GetAbbr() sku.AbbrStore {
	return local.storeAbbr
}
