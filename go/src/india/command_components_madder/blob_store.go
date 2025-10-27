package command_components_madder

import (
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type BlobStore struct{}

func (cmd *BlobStore) MakeBlobStoreFromConfigPath(
	envBlobStore env_repo.BlobStoreEnv,
	basePath string,
	configPath string,
) (blobStore blob_stores.BlobStoreInitialized) {
	var typedConfig blob_store_configs.TypedConfig

	{
		var err error

		if typedConfig, err = triple_hyphen_io.DecodeFromFile(
			blob_store_configs.Coder,
			configPath,
		); err != nil {
			envBlobStore.Cancel(err)
			return blobStore
		}
	}

	blobStore.Config = typedConfig

	blobStore.Path = directory_layout.GetBlobStorePathForCustomPath(
		basePath,
		configPath,
	)

	{
		var err error

		if blobStore.BlobStore, err = blob_stores.MakeBlobStore(
			envBlobStore,
			blobStore.ConfigNamed,
		); err != nil {
			envBlobStore.Cancel(err)
			return blobStore
		}
	}

	return blobStore
}

func (cmd *BlobStore) MakeBlobStoreFromIndexOrConfigPath(
	envBlobStore env_repo.BlobStoreEnv,
	basePath string,
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

		blobStore.Path = directory_layout.GetBlobStorePathForCustomPath(
			basePath,
			blobStoreIndexOrConfigPath,
		)

		{
			var err error

			if blobStore.BlobStore, err = blob_stores.MakeBlobStore(
				envBlobStore,
				blobStore.ConfigNamed,
			); err != nil {
				envBlobStore.Cancel(err)
				return blobStore
			}
		}

		return blobStore
	}

tryBlobStoreIndex:
	return cmd.MakeBlobStoreFromIndex(envBlobStore, blobStoreIndexOrConfigPath)

tryDefaultBlobStore:
	return envBlobStore.GetDefaultBlobStore()
}

func (cmd *BlobStore) MakeBlobStoreFromIndex(
	envBlobStore env_repo.BlobStoreEnv,
	blobStoreIndexString string,
) (blobStore blob_stores.BlobStoreInitialized) {
	var blobStoreIndex int

	{
		var err error

		if blobStoreIndex, err = strconv.Atoi(blobStoreIndexString); err != nil {
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

func (cmd BlobStore) MakeBlobStoresFromIndexesOrAll(
	req command.Request,
	envBlobStore env_repo.BlobStoreEnv,
) []blob_stores.BlobStoreInitialized {
	blobStores := make(
		[]blob_stores.BlobStoreInitialized,
		req.RemainingArgCount(),
	)

	if req.RemainingArgCount() == 0 {
		return envBlobStore.GetBlobStores()
	}

	for i := range blobStores {
		blobStoreIndex := req.PopArg("blob store index")
		blobStores[i] = cmd.MakeBlobStoreFromIndex(envBlobStore, blobStoreIndex)
	}

	return blobStores
}
