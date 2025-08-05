package env_workspace

import (
	"fmt"
	"path/filepath"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/repo_configs"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/workspace_config_blobs"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/store_fs"
	"code.linenisgreat.com/dodder/go/src/mike/store_workspace"
)

type Env interface {
	env_dir.Env
	GetWorkspaceDir() string
	AssertNotTemporary(interfaces.Context)
	AssertNotTemporaryOrOfferToCreate(interfaces.Context)
	IsTemporary() bool
	GetWorkspaceConfig() workspace_config_blobs.Config
	GetWorkspaceConfigFilePath() string
	GetDefaults() repo_configs.Defaults
	CreateWorkspace(workspace_config_blobs.Config) (err error)
	GetStore() *Store

	// TODO identify users of this and reduce / isolate them
	GetStoreFS() *store_fs.Store

	SetWorkspaceTypes(map[string]*Store) (err error)
	SetSupplies(store_workspace.Supplies) (err error)

	Flush() (err error)
}

type Config interface {
	repo_configs.Config
	sku.Config
	interfaces.FileExtensionsGetter
}

func Make(
	envLocal env_local.Env,
	config Config,
	deletedPrinter interfaces.FuncIter[*fd.FD],
	envRepo env_repo.Env,
) (outputEnv *env, err error) {
	outputEnv = &env{
		envRepo:       envRepo,
		Env:           envLocal,
		configMutable: config,
	}

	object := workspace_config_blobs.TypedConfig{
		Type: ids.Type{},
	}

	dir := outputEnv.GetCwd()

	workspaceFile := outputEnv.findWorkspaceFile(dir, env_repo.FileWorkspace)

	if workspaceFile == "" {
		workspaceFile = outputEnv.findWorkspaceFile(
			dir,
			fmt.Sprintf(env_repo.FileWorkspaceTemplate, "zit"),
		)
	}

	if workspaceFile == "" {
		outputEnv.isTemporary = true
		outputEnv.blob = workspace_config_blobs.Temporary{}
	} else {
		if err = workspace_config_blobs.DecodeFromFile(
			&object,
			workspaceFile,
		); err != nil {
			err = errors.BadRequestf("failed to decode `%s`: %w", workspaceFile, err)
			return
		}

		outputEnv.blob = object.Blob
	}

	defaults := outputEnv.configMutable.GetDefaults()

	outputEnv.defaults = repo_configs.DefaultsV1{
		Type: defaults.GetType(),
		Tags: defaults.GetTags(),
	}

	if outputEnv.blob != nil {
		defaults = outputEnv.blob.GetDefaults()

		if newType := defaults.GetType(); !newType.IsEmpty() {
			outputEnv.defaults.Type = newType
		}

		if newTags := defaults.GetTags(); newTags.Len() > 0 {
			outputEnv.defaults.Tags = append(
				outputEnv.defaults.Tags,
				newTags...,
			)
		}
	}

	if outputEnv.isTemporary {
		if outputEnv.dir, err = outputEnv.GetTempLocal().DirTempWithTemplate(
			"workspace-*",
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		// TODO determine this based on the blob
		outputEnv.dir = outputEnv.GetCwd()
	}

	if outputEnv.storeFS, err = store_fs.Make(
		config,
		deletedPrinter,
		config.GetFileExtensions(),
		envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	outputEnv.store.StoreLike = outputEnv.storeFS

	return
}

type env struct {
	envRepo env_repo.Env
	env_local.Env

	isTemporary bool

	// dir is populated on init to either the cwd, or a temporary directory,
	// depending on whether $PWD/.dodder-workspace exists.
	//
	// Later, dir may be set to $PWD/.dodder-workspace by CreateWorkspace
	dir string

	configMutable repo_configs.Config
	blob          workspace_config_blobs.Config
	defaults      repo_configs.DefaultsV1

	storeFS *store_fs.Store
	store   Store
}

func (env *env) findWorkspaceFile(
	dir string,
	name string,
) (found string) {
	// TODO add workspace parent tree height limit?
	for {
		expectedWorkspaceConfigFilePath := filepath.Join(
			dir,
			name,
		)

		if files.Exists(expectedWorkspaceConfigFilePath) {
			found = expectedWorkspaceConfigFilePath
			return
		}

		// if we hit the root, reset to empty so that we trigger the isTemporary
		// path
		if dir == string(filepath.Separator) {
			dir = ""
		}

		dir = filepath.Dir(dir)

		if dir != "." {
			continue
		}

		return
	}
}

func (env *env) GetWorkspaceDir() string {
	return env.dir
}

func (env *env) GetWorkspaceConfigFilePath() string {
	return filepath.Join(env.GetWorkspaceDir(), env_repo.FileWorkspace)
}

func (env *env) AssertNotTemporary(context interfaces.Context) {
	if env.IsTemporary() {
		context.Cancel(ErrNotInWorkspace{env: env})
	}
}

func (env *env) AssertNotTemporaryOrOfferToCreate(context interfaces.Context) {
	if env.IsTemporary() {
		context.Cancel(
			ErrNotInWorkspace{
				env:           env,
				offerToCreate: true,
			})
	}
}

func (env *env) IsTemporary() bool {
	return env.isTemporary
}

func (env *env) GetWorkspaceConfig() workspace_config_blobs.Config {
	return env.blob
}

func (env *env) GetDefaults() repo_configs.Defaults {
	return env.defaults
}

func (env *env) GetStore() *Store {
	return &env.store
}

func (env *env) GetStoreFS() *store_fs.Store {
	return env.storeFS
}

func (env *env) CreateWorkspace(
	blob workspace_config_blobs.Config,
) (err error) {
	env.blob = blob
	tipe := ids.GetOrPanic(ids.TypeTomlWorkspaceConfigV0).Type

	object := workspace_config_blobs.TypedConfig{
		Type: tipe,
		Blob: env.blob,
	}

	env.dir = env.GetCwd()

	if err = workspace_config_blobs.EncodeToFile(
		&object,
		env.GetWorkspaceConfigFilePath(),
	); errors.IsExist(err) {
		err = errors.BadRequestf("workspace already exists")
		return
	} else if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (env *env) SetSupplies(supplies store_workspace.Supplies) (err error) {
	env.store.Supplies = supplies

	if err = env.store.Initialize(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO persist store types and bootstrap based on workspace config
func (env *env) SetWorkspaceTypes(
	stores map[string]*Store,
) (err error) {
	return
}

func (env *env) Flush() (err error) {
	waitGroup := errors.MakeWaitGroupParallel()

	waitGroup.Do(env.store.Flush)

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
