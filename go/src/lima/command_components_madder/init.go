package command_components_madder

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/golf/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/kilo/env_repo"
)

type Init struct{}

func (cmd Init) InitBlobStore(
	ctx interfaces.ActiveContext,
	envBlobStore env_repo.BlobStoreEnv,
	name string,
	config *blob_store_configs.TypedConfig,
) (path directory_layout.BlobStorePath) {
	path = directory_layout.GetBlobStorePath(
		envBlobStore,
		name,
	)

	if err := envBlobStore.MakeDirs(
		filepath.Dir(path.GetBase()),
		filepath.Dir(path.GetConfig()),
	); err != nil {
		envBlobStore.Cancel(err)
		return path
	}

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		config,
		path.GetConfig(),
	); err != nil {
		envBlobStore.Cancel(err)
		return path
	}

	return path
}
