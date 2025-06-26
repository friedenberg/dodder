package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/lima/env_lua"
)

type Stores struct {
	InventoryList InventoryList
	Repo          RepoStore
	Type          Type
	Tag           Tag
}

func MakeStores(
	envRepo env_repo.Env,
	envLua env_lua.Env,
	objectFormat object_inventory_format.Format,
	boxFormat *box_format.BoxTransacted,
) Stores {
	return Stores{
		InventoryList: MakeInventoryStore(envRepo, objectFormat, boxFormat),
		Tag:           MakeTagStore(envRepo, envLua),
		Repo:          MakeRepoStore(envRepo),
		Type:          MakeTypeStore(envRepo),
	}
}

func (a Stores) GetTypeV1() TypedStore[type_blobs.TomlV1, *type_blobs.TomlV1] {
	return a.Type.toml_v1
}
