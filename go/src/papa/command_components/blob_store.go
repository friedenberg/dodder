package command_components

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
	envRepo env_repo.Env,
	blobStoreIndexOrConfigPath string,
) (blobStore env_repo.BlobStoreInitialized) {
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
					envRepo.Cancel(err)
					return
				}
			}
		}

		blobStore.Config = typedConfig.Blob

		{
			var err error

			if blobStore.LocalBlobStore, err = blob_stores.MakeBlobStore(
				envRepo,
				"", // TODO get base path
				typedConfig.Blob,
				envRepo.GetTempLocal(),
			); err != nil {
				envRepo.Cancel(err)
				return
			}
		}

		return
	}

tryBlobStoreIndex:
	{
		var blobStoreIndex int

		{
			var err error

			if blobStoreIndex, err = strconv.Atoi(blobStoreIndexOrConfigPath); err != nil {
				envRepo.Cancel(err)
				return
			}
		}

		blobStores := envRepo.GetBlobStores()

		if len(blobStores)-1 < blobStoreIndex {
			errors.ContextCancelWithBadRequestf(
				envRepo,
				"invalid blob store index: %d. Valid indexes: 0-%d",
				blobStoreIndex,
				len(blobStores)-1,
			)

			return
		}

		blobStore = envRepo.GetBlobStores()[blobStoreIndex]

		return
	}
}
