package command_components_madder

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type BlobStore struct{}

func (cmd *BlobStore) MakeBlobStore(
	envBlobStore env_repo.BlobStoreEnv,
	blobStoreIndexOrConfigPath string,
) (blobStore blob_stores.BlobStoreInitialized) {
	if blobStoreIndexOrConfigPath == "" {
		goto tryDefaultBlobStore
	}

	{
		configPath := blobStoreIndexOrConfigPath
		var typedConfig blob_store_configs.TypedConfig

		{
			var err error

			if typedConfig, err = triple_hyphen_io.DecodeFromFile(
				blob_store_configs.Coder,
				configPath,
			); err != nil {
				if errors.IsNotExist(err) {
					err = nil
					goto tryBlobStoreIndex
				} else {
					envBlobStore.Cancel(err)
					return blobStore
				}
			}
		}

		blobStore.Config = typedConfig

		configNamed := blob_stores.BlobStoreConfigNamed{
			Config: typedConfig,
		}

		configNamed.BasePath, _ = blob_store_configs.GetBasePath(
			typedConfig.Blob,
		)

		{
			var err error

			if blobStore.BlobStore, err = blob_stores.MakeBlobStore(
				envBlobStore,
				configNamed,
				envBlobStore.GetTempLocal(),
			); err != nil {
				envBlobStore.Cancel(err)
				return blobStore
			}
		}

		return blobStore
	}

tryBlobStoreIndex:
	{
		var blobStoreIndex int

		{
			var err error

			if blobStoreIndex, err = strconv.Atoi(blobStoreIndexOrConfigPath); err != nil {
				envBlobStore.Cancel(err)
				return blobStore
			}
		}

		blobStores := envBlobStore.GetBlobStores()

		if len(blobStores)-1 < blobStoreIndex {
			errors.ContextCancelWithBadRequestf(
				envBlobStore,
				"invalid blob store index: %d. Valid indexes: 0-%d",
				blobStoreIndex,
				len(blobStores)-1,
			)

			return blobStore
		}

		blobStore = envBlobStore.GetBlobStores()[blobStoreIndex]

		return blobStore
	}

tryDefaultBlobStore:
	return envBlobStore.GetDefaultBlobStore()
}
