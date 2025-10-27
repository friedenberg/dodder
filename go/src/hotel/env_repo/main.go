package env_repo

import (
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/genesis_configs"
	"code.linenisgreat.com/dodder/go/src/echo/blob_store_configs"
	"code.linenisgreat.com/dodder/go/src/echo/directory_layout"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

const (
	// TODO move to mutable config
	FileWorkspaceTemplate = ".%s-workspace"
	FileWorkspace         = ".dodder-workspace"
)

type Env struct {
	config genesis_configs.TypedConfigPrivate

	lockSmith interfaces.LockSmith

	directory_layout.BlobStore
	directory_layout.Repo

	BlobStoreEnv
}

// TODO stop returning error and just cancel context
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
			return env, err
		}
	}

	xdg := env.GetXDG()

	if env.GetXDG().Data.ActualValue == "" {
		err = errors.Errorf("empty data dir: %#v", env.GetXDG().Data)
		return env, err
	}

	fileConfigPermanent := xdg.Data.MakePath("config-permanent").String()

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
				return env, err
			}
		} else {
			configLoaded = true
		}
	} else {
		if env.config, err = triple_hyphen_io.DecodeFromFile(
			genesis_configs.CoderPrivate,
			fileConfigPermanent,
		); err != nil {
			if errors.IsNotExist(err) {
				err = errors.Wrap(ErrNotInDodderDir{Expected: fileConfigPermanent})
			} else {
				err = errors.Wrap(err)
			}
			return env, err
		} else {
			configLoaded = true
		}
	}

	if env.BlobStore, err = directory_layout.MakeBlobStore(
		env.GetStoreVersion(),
		env.GetXDGForBlobStores(),
	); err != nil {
		err = errors.Wrap(err)
		return env, err
	}

	if env.Repo, err = directory_layout.MakeRepo(
		env.GetStoreVersion(),
		xdg,
	); err != nil {
		err = errors.Wrap(err)
		return env, err
	}

	// TODO fail on pre-existing temp local
	// if files.Exists(s.TempLocal.basePath) {
	// 	err = MakeErrTempAlreadyExists(s.TempLocal.basePath)
	// 	return
	// }

	if err = env.MakeDirsPerms(0o700, env.GetXDG().GetXDGPaths()...); err != nil {
		err = errors.Wrap(err)
		return env, err
	}

	env.lockSmith = file_lock.New(envLocal, env.FileLock(), "repo")

	envVars := env_vars.Make(env)

	env.Must(errors.MakeFuncContextFromFuncErr(envVars.Set))
	env.After(errors.MakeFuncContextFromFuncErr(envVars.Unset))

	if configLoaded {
		env.BlobStoreEnv = MakeBlobStoreEnvFromRepoConfig(
			envLocal,
			env.BlobStore,
			env.GetConfigPrivate().Blob,
		)
	}

	return env, err
}

func (env Env) GetEnv() env_ui.Env {
	return env.Env
}

func (env Env) GetEnvBlobStore() BlobStoreEnv {
	return env.BlobStoreEnv
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
	if err = files.SetAllowUserChangesRecursive(env.DirDataIndex()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = os.RemoveAll(env.DirDataIndex()); err != nil {
		err = errors.Wrapf(err, "failed to remove verzeichnisse dir")
		return err
	}

	if err = env.MakeDirs(env.DirDataIndex()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = env.MakeDirs(env.DirIndexObjects()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = env.MakeDirs(env.DirIndexObjectPointers()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (env Env) DataFileStoreVersion() string {
	return env.GetXDG().Data.MakePath("version").String()
}

func (env Env) GetStoreVersion() store_version.Version {
	if env.config.Blob == nil {
		return store_version.VCurrent
	} else {
		return env.config.Blob.GetStoreVersion()
	}
}

func (env Env) GetInventoryListBlobStore() interfaces.BlobStore {
	storeVersion := env.GetStoreVersion()

	if store_version.LessOrEqual(storeVersion, store_version.V10) {
		return env.getV10OrLessInventoryListBlobStore()
	} else {
		return env.GetDefaultBlobStore()
	}
}

func (env Env) getV10OrLessInventoryListBlobStore() interfaces.BlobStore {
	blob := env.GetConfigPublic().Blob.(interfaces.BlobIOWrapperGetter)

	if store, err := blob_stores.MakeBlobStore(
		env,
		blob_store_configs.ConfigNamed{
			Config: blob_store_configs.TypedConfig{
				Blob: blob.GetBlobIOWrapper().(blob_store_configs.Config),
			},
		},
	); err != nil {
		env.Cancel(err)
		return nil
	} else {
		return store
	}
}

func (env Env) GetBlobStoreById(id int) interfaces.BlobStore {
	return env.blobStores[id]
}
