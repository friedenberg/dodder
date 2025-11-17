package command_components_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blob_store_id"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/golf/fd"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/india/blob_stores"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/juliett/env_repo"
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
		fd.DirBaseOnly(basePath),
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

func (cmd *BlobStore) MakeBlobStoreFromIdOrConfigPath(
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
					goto tryBlobStoreId
				} else {
					envBlobStore.Cancel(err)
					return blobStore
				}
			}
		}

		blobStore.Config = typedConfig

		blobStore.Path = directory_layout.GetBlobStorePathForCustomPath(
			blobStoreIndexOrConfigPath,
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

tryBlobStoreId:
	return cmd.MakeBlobStoreFromIdString(envBlobStore, blobStoreIndexOrConfigPath)

tryDefaultBlobStore:
	return envBlobStore.GetDefaultBlobStore()
}

func (cmd *BlobStore) MakeBlobStoreFromIdString(
	envBlobStore env_repo.BlobStoreEnv,
	blobStoreIdString string,
) (blobStore blob_stores.BlobStoreInitialized) {
	var blobStoreId blob_store_id.Id

	if err := blobStoreId.Set(blobStoreIdString); err != nil {
		envBlobStore.Cancel(err)
		return blobStore
	}

	return envBlobStore.GetBlobStore(blobStoreId)
}

func (cmd BlobStore) MakeBlobStoresFromIdsOrAll(
	req command.Request,
	envBlobStore env_repo.BlobStoreEnv,
) blob_stores.BlobStoreMap {
	blobStores := make(
		blob_stores.BlobStoreMap,
		req.RemainingArgCount(),
	)

	if req.RemainingArgCount() == 0 {
		return envBlobStore.GetBlobStores()
	}

	for range req.RemainingArgCount() {
		blobStoreId := command.PopRequestArg[blob_store_id.Id](
			req.Args,
			"blob store id",
		)

		blobStores[blobStoreId.String()] = envBlobStore.GetBlobStore(
			*blobStoreId,
		)
	}

	return blobStores
}
