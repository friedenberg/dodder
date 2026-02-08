package blob_stores

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/golf/blob_store_configs"
)

type BlobStoreInitialized struct {
	blob_store_configs.ConfigNamed
	interfaces.BlobStore
}

func (blobStoreInitialized BlobStoreInitialized) GetBlobStore() interfaces.BlobStore {
	return blobStoreInitialized.BlobStore
}
