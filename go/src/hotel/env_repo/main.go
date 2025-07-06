package env_repo

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/config_immutable"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/config_immutable_io"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

// TODO move to mutable config
const (
	FileWorkspaceTemplate = ".%s-workspace"
	FileWorkspace         = ".dodder-workspace"
)

type Env struct {
	env_local.Env

	config config_immutable_io.ConfigPrivatedTypedBlob

	readOnlyBlobStorePath string
	lockSmith             interfaces.LockSmith

	interfaces.DirectoryPaths

	local, remote blob_store.LocalBlobStore

	blob_store.CopyingBlobStore
}

func Make(
	envLocal env_local.Env,
	o Options,
) (env Env, err error) {
	env.Env = envLocal
	if o.BasePath == "" {
		o.BasePath = os.Getenv(env_dir.EnvDir)
	}

	if o.BasePath == "" {
		if o.BasePath, err = os.Getwd(); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	env.readOnlyBlobStorePath = o.GetReadOnlyBlobStorePath()

	dp := &directoryV1{}

	if err = dp.init(
		env.GetStoreVersion(),
		env.GetXDG(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	env.DirectoryPaths = dp

	// TODO add support for failing on pre-existing temp local
	// if files.Exists(s.TempLocal.basePath) {
	// 	err = MakeErrTempAlreadyExists(s.TempLocal.basePath)
	// 	return
	// }

	if !o.PermitNoDodderDirectory {
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

	{
		decoder := config_immutable_io.CoderPrivate{}

		if err = decoder.DecodeFromFile(
			&env.config,
			env.FileConfigPermanent(),
		); err != nil {
			errors.Wrap(err)
			return
		}
	}

	if err = env.setupStores(); err != nil {
		errors.Wrap(err)
		return
	}

	return
}

func (env *Env) setupStores() (err error) {
	env.local = env.MakeBlobStore()
	env.CopyingBlobStore = blob_store.MakeCopyingBlobStore(
		env.Env,
		env.local,
		env.remote,
	)

	return
}

func (env Env) GetEnv() env_ui.Env {
	return env.Env
}

func (env Env) GetConfigPublicBlob() config_immutable.ConfigPublic {
	return env.config.ImmutableConfig.GetImmutableConfigPublic()
}

func (env Env) GetConfigPublic() config_immutable_io.ConfigPublicTypedBlob {
	return config_immutable_io.ConfigPublicTypedBlob{
		Type:            env.config.Type,
		ImmutableConfig: env.GetConfigPublicBlob(),
	}
}

func (env Env) GetConfigPrivate() config_immutable_io.ConfigPrivatedTypedBlob {
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

func (env Env) GetStoreVersion() interfaces.StoreVersion {
	if env.config.ImmutableConfig == nil {
		return store_version.VCurrent
	} else {
		return env.config.ImmutableConfig.GetStoreVersion()
	}
}

func (env Env) Mover() (*env_dir.Mover, error) {
	return env.local.Mover()
}

func (env Env) MakeBlobStore() blob_store.LocalBlobStore {
	return blob_store.MakeShardedFilesStore(
		env.DirBlobs(),
		env_dir.MakeConfigFromImmutableBlobConfig(
			env.GetConfigPrivate().ImmutableConfig.GetBlobStoreConfigImmutable(),
		),
		env.GetTempLocal(),
	)
}
