package local_working_copy

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/config_immutable_io"
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

func (u *Repo) GetEnv() env_ui.Env {
	return u
}

func (u *Repo) GetImmutableConfigPublic() config_immutable_io.ConfigPublicTypedBlob {
	return u.GetEnvRepo().GetConfigPublic()
}

func (u *Repo) GetImmutableConfigPrivate() config_immutable_io.ConfigPrivateTypedBlob {
	return u.GetEnvRepo().GetConfigPrivate()
}

func (repo *Repo) GetEnvLocal() env_local.Env {
	return repo
}

func (repo *Repo) GetEnvWorkspace() env_workspace.Env {
	return repo.envWorkspace
}

func (u *Repo) GetEnvLua() env_lua.Env {
	return u.envLua
}

func (u *Repo) GetTime() time.Time {
	return time.Now()
}

func (u *Repo) GetConfig() store_config.Store {
	return u.config
}

func (u *Repo) GetDormantIndex() *dormant_index.Index {
	return &u.dormantIndex
}

func (u *Repo) GetEnvRepo() env_repo.Env {
	return u.envRepo
}

func (u *Repo) GetTypedInventoryListBlobStore() typed_blob_store.InventoryList {
	return u.typedBlobStore.InventoryList
}

func (u *Repo) GetBlobStore() interfaces.BlobStore {
	return u.GetEnvRepo()
}

func (repo *Repo) GetObjectStore() sku.ObjectStore {
	return &repo.store
}

func (u *Repo) GetInventoryListStore() sku.InventoryListStore {
	return u.GetStore().GetInventoryListStore()
}

func (u *Repo) GetStore() *store.Store {
	return &u.store
}

func (repo *Repo) GetAbbr() sku.AbbrStore {
	return repo.storeAbbr
}
