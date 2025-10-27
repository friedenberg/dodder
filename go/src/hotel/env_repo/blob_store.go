package env_repo

import (
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

type BlobStoreEnv struct {
	directory_layout.BlobStore
	env_local.Env

	blobStoreDefaultIndex int

	// TODO switch to implementing LocalBlobStore directly and writing to all of
	// the defined blob stores instead of having a default
	// TODO switch to primary blob store and others, and add support for v10
	// directory layout
	blobStores []blob_stores.BlobStoreInitialized
}

func MakeBlobStoreEnvFromRepoConfig(
	envLocal env_local.Env,
	directoryLayout directory_layout.BlobStore,
	config genesis_configs.ConfigPrivate,
) BlobStoreEnv {
	env := BlobStoreEnv{
		Env:       envLocal,
		BlobStore: directoryLayout,
	}

	env.setupStoresFromRepoConfig(config)

	return env
}

func (env *BlobStoreEnv) setupStoresFromRepoConfig(
	config genesis_configs.ConfigPrivate,
) {
	env.blobStores = blob_stores.MakeBlobStoresFromRepoConfig(
		env,
		env,
		config,
		env.BlobStore,
	)
}

func MakeBlobStoreEnv(
	envLocal env_local.Env,
) BlobStoreEnv {
	env := BlobStoreEnv{
		Env: envLocal,
	}

	{
		var err error

		if env.BlobStore, err = directory_layout.MakeRepo(
			store_version.VCurrent,
			envLocal.GetXDG(),
		); err != nil {
			envLocal.Cancel(err)
			return env
		}
	}

	env.setupStores()

	return env
}

func (env *BlobStoreEnv) setupStores() {
	env.blobStores = blob_stores.MakeBlobStores(
		env,
		env,
		env.BlobStore,
	)
}

func (env BlobStoreEnv) GetDefaultBlobStore() blob_stores.BlobStoreInitialized {
	if len(env.blobStores) == 0 {
		panic("calling GetDefaultBlobStore without any initialized blob stores")
	}

	return env.blobStores[env.blobStoreDefaultIndex]
}

func (env BlobStoreEnv) GetBlobStores() []blob_stores.BlobStoreInitialized {
	blobStores := make([]blob_stores.BlobStoreInitialized, len(env.blobStores))
	copy(blobStores, env.blobStores)
	return blobStores
}

func (env *BlobStoreEnv) writeBlobStoreConfig(
	bigBang BigBang,
	directoryLayout directory_layout.BlobStore,
) {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		// the immutable config contains the only blob stores's config
		return
	}

	blobStoreConfigPath := directory_layout.GetDefaultBlobStore(
		directoryLayout,
	).Config

	blobStoreConfigDir := filepath.Dir(blobStoreConfigPath)

	if err := env.MakeDirs(blobStoreConfigDir); err != nil {
		env.Cancel(err)
		return
	}

	blobStoreConfig := bigBang.TypedBlobStoreConfig

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		&blob_store_configs.TypedConfig{
			Type: blobStoreConfig.Type,
			Blob: blobStoreConfig.Blob,
		},
		blobStoreConfigPath,
	); err != nil {
		env.Cancel(err)
		return
	}
}
