package repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
	"code.linenisgreat.com/dodder/go/src/kilo/inventory_list_coders"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
)

// TODO explore permissions for who can read / write from the inventory list
// store
type Repo interface {
	GetEnv() env_ui.Env
	GetImmutableConfigPublic() genesis_configs.ConfigPublic
	GetImmutableConfigPublicType() ids.Type
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
		builderOptions query.BuilderOption,
		externalQueryOptions sku.ExternalQueryOptions,
		args ...string,
	) (qg *query.Query, err error)

	MakeInventoryList(
		qg *query.Query,
	) (list *sku.ListTransacted, err error)

	// TODO replace with WorkingCopy
	PullQueryGroupFromRemote(
		remote Repo,
		qg *query.Query,
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
