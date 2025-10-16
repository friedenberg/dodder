package env_repo

import (
	"fmt"
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
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
		Env: envLocal,
	}

	directoryLayout := &directoryLayoutV2{}

	if err := directoryLayout.initDirectoryLayout(envLocal.GetXDG()); err != nil {
		envLocal.Cancel(err)
		return env
	}

	env.BlobStoreDirectoryLayout = directoryLayout

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

func (env *BlobStoreEnv) writeBlobStoreConfig(
	bigBang BigBang,
	directoryLayout interfaces.BlobStoreDirectoryLayout,
) {
	if store_version.IsCurrentVersionLessOrEqualToV10() {
		// the immutable config contains the only blob stores's config
		return
	}

	blobStoreConfig := bigBang.TypedBlobStoreConfig

	if config, ok := blobStoreConfig.Blob.(blob_store_configs.ConfigLocalMutable); ok {
		config.SetBasePath(
			interfaces.DirectoryLayoutDirBlobStore(directoryLayout, strconv.Itoa(0)),
		)
	}

	if err := triple_hyphen_io.EncodeToFile(
		blob_store_configs.Coder,
		&blob_store_configs.TypedConfig{
			Type: blobStoreConfig.Type,
			Blob: blobStoreConfig.Blob,
		},
		directoryLayout.DirBlobStoreConfigs(
			fmt.Sprintf("%d-default.%s", 0, FileNameBlobStoreConfig),
		),
	); err != nil {
		env.Cancel(err)
		return
	}
}
