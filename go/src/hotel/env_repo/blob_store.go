package env_repo

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

type BlobStoreEnv struct {
	interfaces.BlobStoreDirectoryLayout
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
	directoryLayout interfaces.BlobStoreDirectoryLayout,
	config genesis_configs.ConfigPrivate,
) BlobStoreEnv {
	env := BlobStoreEnv{
		Env:                      envLocal,
		BlobStoreDirectoryLayout: directoryLayout,
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
		env.BlobStoreDirectoryLayout,
	)
}

func MakeBlobStoreEnv(
	envLocal env_local.Env,
) BlobStoreEnv {
	env := BlobStoreEnv{
		Env:                      envLocal,
		BlobStoreDirectoryLayout: &directoryLayoutV2{},
	}

	env.setupStores()

	return env
}

func (env *BlobStoreEnv) setupStores() {
	env.blobStores = blob_stores.MakeBlobStores(
		env,
		env,
		env.BlobStoreDirectoryLayout,
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
