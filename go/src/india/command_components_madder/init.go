package command_components_madder

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type Init struct{}

func (cmd Init) InitBlobStore(
	ctx interfaces.ActiveContext,
	envBlobStore env_repo.BlobStoreEnv,
	name string,
	config *triple_hyphen_io.TypedBlob[blob_store_configs.Config],
) (pathConfig string) {
	blobStoreCount := len(envBlobStore.GetBlobStores())

	dir, target := directory_layout.GetBlobStoreConfigPath(
		envBlobStore,
		blobStoreCount,
		name,
	)

	if err := envBlobStore.MakeDir(dir); err != nil {
		envBlobStore.Cancel(err)
		return pathConfig
	}

	pathConfig = filepath.Join(dir, target)

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		config,
		pathConfig,
	); err != nil {
		envBlobStore.Cancel(err)
		return pathConfig
	}

	return pathConfig
}
