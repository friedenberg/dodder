package repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

// TODO explore permissions for who can read / write from the inventory list
// store
type Repo interface {
	GetEnv() env_ui.Env
	GetImmutableConfigPublic() config_immutable.Public
	GetImmutableConfigPublicType() ids.Type
	GetBlobStore() interfaces.BlobStore
	GetObjectStore() sku.ObjectStore
	GetTypedInventoryListBlobStore() typed_blob_store.InventoryList
	GetInventoryListStore() sku.InventoryListStore

	MakeImporter(
		options sku.ImporterOptions,
		storeOptions sku.StoreOptions,
	) sku.Importer

	// TODO switch to seq
	ImportList(
		list *sku.List,
		i sku.Importer,
	) (err error)
}

type WorkingCopy interface {
	Repo

	// MakeQueryGroup(
	// 	builderOptions query.BuilderOptions,
	// 	args ...string,
	// ) (qg *query.Group, err error)

	MakeExternalQueryGroup(
		builderOptions query.BuilderOption,
		externalQueryOptions sku.ExternalQueryOptions,
		args ...string,
	) (qg *query.Query, err error)

	MakeInventoryList(
		qg *query.Query,
	) (list *sku.List, err error)

	PullQueryGroupFromRemote(
		remote Repo,
		qg *query.Query,
		options RemoteTransferOptions,
	) (err error)

	ReadObjectHistory(
		oid *ids.ObjectId,
	) (skus []*sku.Transacted, err error)
}
