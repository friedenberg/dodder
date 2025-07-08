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
	"code.linenisgreat.com/dodder/go/src/echo/triple_hyphen_io2"
	"code.linenisgreat.com/dodder/go/src/foxtrot/builtin_types"
	"code.linenisgreat.com/dodder/go/src/golf/repo_config_blobs"
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
	AssertNotTemporary(errors.Context)
	AssertNotTemporaryOrOfferToCreate(errors.Context)
	IsTemporary() bool
	GetWorkspaceConfig() workspace_config_blobs.Blob
	GetDefaults() repo_config_blobs.Defaults
	CreateWorkspace(workspace_config_blobs.Blob) (err error)
	DeleteWorkspace() (err error)
	GetStore() *Store

	// TODO identify users of this and reduce / isolate them
	GetStoreFS() *store_fs.Store

	SetWorkspaceTypes(map[string]*Store) (err error)
	SetSupplies(store_workspace.Supplies) (err error)

	Flush() (err error)
}

type Config interface {
	repo_config_blobs.Getter
	sku.Config
	interfaces.FileExtensionsGetter
}

func Make(
	envLocal env_local.Env,
	config Config,
	deletedPrinter interfaces.FuncIter[*fd.FD],
	envRepo env_repo.Env,
) (out *env, err error) {
	out = &env{
		envRepo:       envRepo,
		Env:           envLocal,
		configMutable: config.GetMutableConfig(),
	}

	object := triple_hyphen_io2.TypedBlob[*workspace_config_blobs.Blob]{
		Type: ids.Type{},
	}

	dir := out.GetCwd()

	workspaceFile := out.findWorkspaceFile(dir, env_repo.FileWorkspace)

	if workspaceFile == "" {
		workspaceFile = out.findWorkspaceFile(
			dir,
			fmt.Sprintf(env_repo.FileWorkspaceTemplate, "zit"),
		)
	}

	if workspaceFile == "" {
		out.isTemporary = true
	} else {
		if err = workspace_config_blobs.DecodeFromFile(
			&object,
			workspaceFile,
		); err != nil {
			err = errors.BadRequestf("failed to decode `%s`: %w", workspaceFile, err)
			return
		}

		out.blob = *object.Blob
	}

	defaults := out.configMutable.GetDefaults()

	out.defaults = repo_config_blobs.DefaultsV1{
		Type: defaults.GetType(),
		Tags: defaults.GetTags(),
	}

	if out.blob != nil {
		defaults = out.blob.GetDefaults()

		if newType := defaults.GetType(); !newType.IsEmpty() {
			out.defaults.Type = newType
		}

		if newTags := defaults.GetTags(); newTags.Len() > 0 {
			out.defaults.Tags = append(out.defaults.Tags, newTags...)
		}
	}

	if out.isTemporary {
		if out.dir, err = out.GetTempLocal().DirTempWithTemplate(
			"workspace-*",
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		out.dir = out.GetCwd()
	}

	if out.storeFS, err = store_fs.Make(
		config,
		deletedPrinter,
		config.GetFileExtensions(),
		envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	out.store.StoreLike = out.storeFS

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

	configMutable repo_config_blobs.Blob
	blob          workspace_config_blobs.Blob
	defaults      repo_config_blobs.DefaultsV1

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

func (env *env) tryLoad(
	path string,
	object *triple_hyphen_io2.TypedBlob[*workspace_config_blobs.Blob],
) (err error) {
	if err = workspace_config_blobs.DecodeFromFile(
		object,
		path,
	); err != nil {
		err = errors.BadRequestPrefix("failed to decode `.dodder-workspace`", err)
		return
	}

	env.blob = *object.Blob

	return
}

func (env *env) GetWorkspaceDir() string {
	return env.dir
}

func (env *env) GetWorkspaceConfigFilePath() string {
	return filepath.Join(env.GetWorkspaceDir(), env_repo.FileWorkspace)
}

func (env *env) AssertNotTemporary(context errors.Context) {
	if env.IsTemporary() {
		context.CancelWithError(ErrNotInWorkspace{env: env})
	}
}

func (env *env) AssertNotTemporaryOrOfferToCreate(context errors.Context) {
	if env.IsTemporary() {
		context.CancelWithError(
			ErrNotInWorkspace{
				env:           env,
				offerToCreate: true,
			},
		)
	}
}

func (env *env) IsTemporary() bool {
	return env.isTemporary
}

func (env *env) GetWorkspaceConfig() workspace_config_blobs.Blob {
	return env.blob
}

func (env *env) GetDefaults() repo_config_blobs.Defaults {
	return env.defaults
}

func (env *env) GetStore() *Store {
	return &env.store
}

func (env *env) GetStoreFS() *store_fs.Store {
	return env.storeFS
}

func (env *env) CreateWorkspace(blob workspace_config_blobs.Blob) (err error) {
	env.blob = blob
	tipe := builtin_types.GetOrPanic(builtin_types.WorkspaceConfigTypeTomlV0).Type

	object := triple_hyphen_io2.TypedBlob[*workspace_config_blobs.Blob]{
		Type: tipe,
		Blob: &env.blob,
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

func (env *env) DeleteWorkspace() (err error) {
	if err = env.Delete(env.GetWorkspaceConfigFilePath()); errors.IsNotExist(err) {
		err = nil
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
