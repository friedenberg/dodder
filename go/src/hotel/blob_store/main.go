package blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_config"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
)

func MakeBlobStore(
	ctx errors.Context,
	basePath string,
	config blob_store_config.Config,
	tempFS env_dir.TemporaryFS,
) interfaces.LocalBlobStore {
	switch config := config.(type) {
	default:
		ctx.CancelWithErrorf("unsupported blob store config: %T", config)
		return nil

	case gitLikeBucketedConfig:
		return makeGitLikeBucketedStore(basePath, config, tempFS)
	}
}
