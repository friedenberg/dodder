package repo

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_ui"
	"code.linenisgreat.com/dodder/go/src/india/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/juliett/blob_stores"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/mike/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/oscar/queries"
	"code.linenisgreat.com/dodder/go/src/romeo/env_workspace"
)

// TODO explore permissions for who can read / write from the inventory list
// store
type Repo interface {
	GetEnv() env_ui.Env
	GetImmutableConfigPublic() genesis_configs.ConfigPublic
	GetImmutableConfigPublicType() ids.TypeStruct
	GetBlobStore() blob_stores.BlobStoreInitialized
	GetObjectStore() sku.RepoStore
	GetInventoryListCoderCloset() inventory_list_coders.Closet
	GetInventoryListStore() sku.InventoryListStore

	MakeImporter(
		options ImporterOptions,
		storeOptions sku.StoreOptions,
	) Importer

	ImportSeq(
		interfaces.SeqError[*sku.Transacted],
		Importer,
	) error

	MakeExternalQueryGroup(
		builderOptions queries.BuilderOption,
		externalQueryOptions sku.ExternalQueryOptions,
		args ...string,
	) (qg *queries.Query, err error)

	MakeInventoryList(
		qg *queries.Query,
	) (list *sku.HeapTransacted, err error)

	// TODO replace with WorkingCopy
	PullQueryGroupFromRemote(
		remote Repo,
		qg *queries.Query,
		options ImporterOptions,
	) (err error)

	ReadObjectHistory(
		oid *ids.ObjectId,
	) (skus []*sku.Transacted, err error)
}

type LocalRepo interface {
	Repo

	GetEnvRepo() env_repo.Env
	GetImmutableConfigPrivate() genesis_configs.TypedConfigPrivate

	Lock() error
	Unlock() error

	GetEnvWorkspace() env_workspace.Env
}
