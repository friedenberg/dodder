package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/type_blobs"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/lima/env_lua"
)

type Stores struct {
	InventoryList inventory_list_coders.Closet
	Repo          RepoStore
	Type          Type
	Tag           Tag
}

func MakeStores(
	envRepo env_repo.Env,
	envLua env_lua.Env,
	boxFormat *box_format.BoxTransacted,
) Stores {
	return Stores{
		InventoryList: inventory_list_coders.MakeCloset(
			envRepo,
			boxFormat,
		),
		Tag:  MakeTagStore(envRepo, envLua),
		Repo: MakeRepoStore(envRepo),
		Type: MakeTypeStore(envRepo),
	}
}

func (stores Stores) GetTypeV1() TypedStore[type_blobs.TomlV1, *type_blobs.TomlV1] {
	return stores.Type.toml_v1
}
