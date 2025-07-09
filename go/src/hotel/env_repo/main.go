package env_repo

import (
	"os"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/env_vars"
	"code.linenisgreat.com/dodder/go/src/bravo/todo"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/store_version"
	"code.linenisgreat.com/dodder/go/src/delta/file_lock"
	"code.linenisgreat.com/dodder/go/src/delta/xdg"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/golf/genesis_config_io"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_store"
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

type Env struct {
	env_local.Env

	config genesis_config_io.PrivateTypedBlob

	readOnlyBlobStorePath string
	lockSmith             interfaces.LockSmith

	directoryPaths

	local, remote blob_store.LocalBlobStore

	blob_store.CopyingBlobStore
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

	// TODO add support for failing on pre-existing temp local
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

	{
		decoder := genesis_config_io.CoderPrivate{}

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
	env.local = env.GetDefaultBlobStore()
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

func (env Env) GetConfigPublic() genesis_config_io.PublicTypedBlob {
	return genesis_config_io.PublicTypedBlob{
		Type: env.config.Type,
		Blob: env.config.Blob.GetImmutableConfigPublic(),
	}
}

func (env Env) GetConfigPrivate() genesis_config_io.PrivateTypedBlob {
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

func (env Env) Mover() (*env_dir.Mover, error) {
	return env.local.Mover()
}

func (env Env) GetDefaultBlobStore() blob_store.LocalBlobStore {
	// TODO use default blob store ref from config and initialize a blob store
	// TODO depending on store version, read immutable config from blob store path
	// or from config
	return blob_store.MakeShardedFilesStore(
		env.DirFirstBlobStoreBlobs(),
		env_dir.MakeConfigFromImmutableBlobConfig(
			env.GetConfigPublic().Blob.GetBlobStoreConfigImmutable(),
		),
		env.GetTempLocal(),
	)
}

func (env Env) GetBlobStoreById(id string) interfaces.BlobStore {
	panic(todo.Implement())
}

// func (env Env) MakeBlobStoreFromConfig(
// 	blobStoreConfig config_immutable.BlobStoreConfig,
// ) blob_store.LocalBlobStore {
// 	return blob_store.MakeShardedFilesStore(
// 		env.DirBlobs(),
// 		env_dir.MakeConfigFromImmutableBlobConfig(blobStoreConfig),
// 		env.GetTempLocal(),
// 	)
// }

func (env Env) MakeBlobStoreFromName(
	name string,
) (interfaces.BlobStore, error) {
	return nil, todo.Implement()
}
