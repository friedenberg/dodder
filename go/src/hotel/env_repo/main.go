package env_repo

import (
	"os"
	"path/filepath"
	"strconv"

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
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

const (
	// TODO move to mutable config
	FileWorkspaceTemplate = ".%s-workspace"
	FileWorkspace         = ".dodder-workspace"

	FileNameBlobStoreConfig = "dodder-blob_store-config"
)

type directoryLayout interface {
	interfaces.DirectoryLayout
	initDirectoryLayout(xdg.XDG) error
}

type BlobStoreInitialized struct {
	Name     string
	BasePath string
	blob_store_configs.Config
	interfaces.BlobStore
}

type Env struct {
	env_local.Env

	config genesis_configs.TypedConfigPrivate

	lockSmith interfaces.LockSmith

	directoryLayout

	blobStoreDefaultIndex int

	// TODO switch to implementing LocalBlobStore directly and writing to all of
	// the defined blob stores instead of having a default
	// TODO switch to primary blob store and others, and add support for v10
	// directory layout
	blobStores []BlobStoreInitialized
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

	xdg := env.GetXDG()
	fileConfigPermanent := filepath.Join(xdg.Data, "config-permanent")

	var configLoaded bool

	if options.PermitNoDodderDirectory {
		if env.config, err = triple_hyphen_io.DecodeFromFile(
			genesis_configs.CoderPrivate,
			fileConfigPermanent,
		); err != nil {
			if errors.IsNotExist(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
				return
			}
		} else {
			configLoaded = true
		}
	} else {
		if env.config, err = triple_hyphen_io.DecodeFromFile(
			genesis_configs.CoderPrivate,
			fileConfigPermanent,
		); err != nil {
			err = errors.Wrap(ErrNotInDodderDir{})
			return
		} else {
			configLoaded = true
		}
	}

	if env.GetStoreVersion().LessOrEqual(store_version.V10) {
		env.directoryLayout = &directoryLayoutV1{}
	} else {
		env.directoryLayout = &directoryLayoutV2{}
	}

	if err = env.initDirectoryLayout(xdg); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO fail on pre-existing temp local
	// if files.Exists(s.TempLocal.basePath) {
	// 	err = MakeErrTempAlreadyExists(s.TempLocal.basePath)
	// 	return
	// }

	if env.DirDodder() == "" {
		panic("empty dir dodder")
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

	if configLoaded {
		env.setupStores()
	}

	return
}

func (env *Env) setupStores() {
	if store_version.LessOrEqual(env.GetStoreVersion(), store_version.V10) {
		env.blobStores = make([]BlobStoreInitialized, 1)
		blob := env.GetConfigPublic().Blob.(interfaces.BlobIOWrapperGetter)
		env.blobStores[0].Name = "0-default"
		env.blobStores[0].Config = blob.GetBlobIOWrapper().(blob_store_configs.Config)
		env.blobStores[0].BasePath = env.DirBlobStores("blobs")
	} else {
		var configPaths []string

		{
			var err error

			if configPaths, err = files.DirNames(
				filepath.Join(env.DirBlobStoreConfigs()),
			); err != nil {
				env.Cancel(err)
				return
			}
		}

		env.blobStores = make([]BlobStoreInitialized, len(configPaths))

		for i, configPath := range configPaths {
			env.blobStores[i].Name = fd.FileNameSansExt(configPath)
			env.blobStores[i].BasePath = env.DirBlobStores(strconv.Itoa(i))

			if typedConfig, err := triple_hyphen_io.DecodeFromFile(
				blob_store_configs.Coder,
				configPath,
			); err != nil {
				env.Cancel(err)
				return
			} else {
				env.blobStores[i].Config = typedConfig.Blob
			}
		}
	}

	for i, blobStore := range env.blobStores {
		var err error

		// TODO use sha of config to determine blob store base path
		if env.blobStores[i].BlobStore, err = blob_stores.MakeBlobStore(
			env,
			blobStore.BasePath,
			blobStore.Config,
			env.GetTempLocal(),
		); err != nil {
			env.Cancel(err)
			return
		}
	}
}

func (env Env) GetEnv() env_ui.Env {
	return env.Env
}

func (env Env) GetConfigPublic() genesis_configs.TypedConfigPublic {
	return genesis_configs.TypedConfigPublic{
		Type: env.config.Type,
		Blob: env.config.Blob.GetGenesisConfigPublic(),
	}
}

func (env Env) GetConfigPrivate() genesis_configs.TypedConfigPrivate {
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

func (env Env) GetDefaultBlobStore() BlobStoreInitialized {
	if len(env.blobStores) == 0 {
		panic("calling GetDefaultBlobStore without any initialized blob stores")
	}

	return env.blobStores[env.blobStoreDefaultIndex]
}

func (env Env) GetBlobStores() []BlobStoreInitialized {
	blobStores := make([]BlobStoreInitialized, len(env.blobStores))
	copy(blobStores, env.blobStores)
	return blobStores
}

func (env Env) GetInventoryListBlobStore() interfaces.BlobStore {
	storeVersion := env.GetStoreVersion()

	if store_version.LessOrEqual(storeVersion, store_version.V10) {
		blob := env.GetConfigPublic().Blob.(interfaces.BlobIOWrapperGetter)

		if store, err := blob_stores.MakeBlobStore(
			env,
			env.DirFirstBlobStoreInventoryLists(),
			blob.GetBlobIOWrapper().(blob_store_configs.Config),
			env.GetTempLocal(),
		); err != nil {
			env.Cancel(err)
			return nil
		} else {
			return store
		}
	} else {
		return env.GetDefaultBlobStore()
	}
}

func (env Env) GetBlobStoreById(id int) interfaces.BlobStore {
	return env.blobStores[id]
}
