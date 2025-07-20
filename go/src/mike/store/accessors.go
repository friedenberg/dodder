package store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/thyme"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/zettel_id_index"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/stream_index"
	"code.linenisgreat.com/dodder/go/src/lima/inventory_list_store"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

func (store *Store) GetTypedBlobStore() typed_blob_store.Stores {
	return store.typedBlobStore
}

func (store *Store) GetEnnui() object_probe_index.Index {
	return nil
}

func (store *Store) GetProtoZettel() sku.Proto {
	return store.protoZettel
}

func (store *Store) GetPersistentMetadataFormat() object_inventory_format.Format {
	return store.persistentObjectFormat
}

func (store *Store) GetTime() thyme.Time {
	return thyme.Now()
}

func (store *Store) GetTai() ids.Tai {
	return ids.NowTai()
}

func (store *Store) GetInventoryListStore() *inventory_list_store.Store {
	return &store.inventoryListStore
}

func (store *Store) GetAbbrStore() sku.AbbrStore {
	return store.Abbr
}

func (store *Store) GetZettelIdIndex() zettel_id_index.Index {
	return store.zettelIdIndex
}

func (store *Store) GetEnvRepo() env_repo.Env {
	return store.envRepo
}

func (store *Store) GetConfigStore() store_config.Store {
	return store.storeConfig
}

func (store *Store) GetConfigStoreMutable() store_config.StoreMutable {
	return store.storeConfig
}

func (store *Store) GetStreamIndex() *stream_index.Index {
	return store.streamIndex
}

func (store *Store) GetConfigBlobFormat() interfaces.CoderReadWriter[*repo_configs.TypedBlob] {
	return store.configBlobCoder
}
