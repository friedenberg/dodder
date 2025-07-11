package env_repo

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

// TODO move to mutable config
const (
	FileWorkspaceTemplate = ".%s-workspace"
	FileWorkspace         = ".dodder-workspace"
)

type directoryPaths interface {
	interfaces.DirectoryPaths
	init(interfaces.StoreVersion, xdg.XDG) error
}

type BlobStoreWithConfig struct {
	blob_store_configs.Config
	interfaces.LocalBlobStore
}

type Env struct {
	env_local.Env

	config genesis_configs.PrivateTypedBlob

	readOnlyBlobStorePath string
	lockSmith             interfaces.LockSmith

	directoryPaths

	blobStoreDefaultIndex int

	// TODO switch to implementing LocalBlobStore directly and writing to all of
	// the defined blob stores instead of having a default
	blobStores []BlobStoreWithConfig
}

func Make(
	envLocal env_local.Env,
	options Options,
) (env Env, err error) {
	env.Env = envLocal

	if options.BasePath == "" {
		options.BasePath = os.Getenv(env_dir.EnvDir)
	}

	if options.BasePath == "" {
		if options.BasePath, err = os.Getwd(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	env.readOnlyBlobStorePath = options.GetReadOnlyBlobStorePath()

	if env.GetStoreVersion().LessOrEqual(store_version.V10) {
		env.directoryPaths = &directoryV1{}
	} else {
		env.directoryPaths = &directoryV2{}
	}

	if err = env.directoryPaths.init(
		env.GetStoreVersion(),
		env.GetXDG(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO fail on pre-existing temp local
	// if files.Exists(s.TempLocal.basePath) {
	// 	err = MakeErrTempAlreadyExists(s.TempLocal.basePath)
	// 	return
	// }

	if !options.PermitNoDodderDirectory {
		if ok := files.Exists(env.DirDodder()); !ok {
			err = errors.Wrap(ErrNotInDodderDir{Expected: env.DirDodder()})
			return
		}
	}

	if err = env.MakeDirPerms(0o700, env.GetXDG().GetXDGPaths()...); err != nil {
		err = errors.Wrap(err)
		return
	}

	env.lockSmith = file_lock.New(envLocal, env.FileLock(), "repo")

	envVars := env_vars.Make(env)

	for key, value := range envVars {
		if err = os.Setenv(key, value); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	env.config = triple_hyphen_io.DecodeFromFile(
		env,
		genesis_configs.CoderPrivate,
		env.FileConfigPermanent(),
		true,
	)

	env.setupStores()

	return
}

func (env *Env) setupStores() {
	if store_version.LessOrEqual(env.GetStoreVersion(), store_version.V10) {
		env.blobStores = make([]BlobStoreWithConfig, 1)
		env.blobStores[0].Config = env.GetConfigPublic().Blob.GetBlobStoreConfigImmutable()
	} else {
		var configPaths []string

		{
			var err error

			// TODO consider just iterating and using ErrNotExist instead
			if configPaths, err = filepath.Glob(
				filepath.Join(env.DirBlobStores(), "*", "config.toml"),
			); err != nil {
				env.CancelWithError(err)
			}
		}

		env.blobStores = make([]BlobStoreWithConfig, len(configPaths))

		for i, configPath := range configPaths {
			env.blobStores[i].Config = triple_hyphen_io.DecodeFromFile(
				env,
				blob_store_configs.Coder,
				configPath,
				false,
			).Blob
		}
	}

	for i, blobStore := range env.blobStores {
		env.blobStores[i].LocalBlobStore = blob_stores.MakeBlobStore(
			env,
			env.DirFirstBlobStoreBlobs(),
			blobStore.Config,
			env.GetTempLocal(),
		)
	}
}

func (env Env) GetEnv() env_ui.Env {
	return env.Env
}

func (env Env) GetConfigPublic() genesis_configs.PublicTypedBlob {
	return genesis_configs.PublicTypedBlob{
		Type: env.config.Type,
		Blob: env.config.Blob.GetImmutableConfigPublic(),
	}
}

func (env Env) GetConfigPrivate() genesis_configs.PrivateTypedBlob {
	return env.config
}

func (env Env) GetLockSmith() interfaces.LockSmith {
	return env.lockSmith
}

func stringSliceJoin(s string, vs []string) []string {
	return append([]string{s}, vs...)
}

func (env Env) ResetCache() (err error) {
	if err = files.SetAllowUserChangesRecursive(env.DirCache()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = os.RemoveAll(env.DirCache()); err != nil {
		err = errors.Wrapf(err, "failed to remove verzeichnisse dir")
		return
	}

	if err = env.MakeDir(env.DirCache()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = env.MakeDir(env.DirCacheObjects()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = env.MakeDir(env.DirCacheObjectPointers()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (env Env) DataFileStoreVersion() string {
	return filepath.Join(env.GetXDG().Data, "version")
}

func (env Env) GetStoreVersion() store_version.Version {
	if env.config.Blob == nil {
		return store_version.VCurrent
	} else {
		return env.config.Blob.GetStoreVersion()
	}
}

func (env Env) GetDefaultBlobStore() BlobStoreWithConfig {
	if len(env.blobStores) == 0 {
		panic("calling GetDefaultBlobStore without any initialized blob stores")
	}

	return env.blobStores[env.blobStoreDefaultIndex]
}

func (env Env) GetBlobStores() []BlobStoreWithConfig {
	blobStores := make([]BlobStoreWithConfig, len(env.blobStores))
	copy(blobStores, env.blobStores)
	return blobStores
}

func (env Env) GetInventoryListBlobStore() interfaces.LocalBlobStore {
	storeVersion := env.GetStoreVersion()

	if store_version.LessOrEqual(storeVersion, store_version.V10) {
		return blob_stores.MakeBlobStore(
			env,
			env.DirFirstBlobStoreInventoryLists(),
			env.GetConfigPrivate().Blob.GetBlobStoreConfigImmutable(),
			env.GetTempLocal(),
		)
	} else {
		return env.GetDefaultBlobStore()
	}
}

func (env Env) GetBlobStoreById(id int) interfaces.BlobStore {
	return env.blobStores[id]
}
