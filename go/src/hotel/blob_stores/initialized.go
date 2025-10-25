package blob_stores

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
)

type BlobStoreInitialized struct {
	blob_store_configs.ConfigNamed
	interfaces.BlobStore
}

func (blobStoreInitialized BlobStoreInitialized) GetBlobStore() interfaces.BlobStore {
	return blobStoreInitialized.BlobStore
}
