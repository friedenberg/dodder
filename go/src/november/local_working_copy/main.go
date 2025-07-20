package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/repo_type"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/store_workspace"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/hotel/object_inventory_format"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/box_format"
	"code.linenisgreat.com/dodder/go/src/kilo/dormant_index"
	"code.linenisgreat.com/dodder/go/src/kilo/env_workspace"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/kilo/store_abbr"
	"code.linenisgreat.com/dodder/go/src/lima/env_lua"
	"code.linenisgreat.com/dodder/go/src/lima/store_browser"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
	"code.linenisgreat.com/dodder/go/src/mike/env_box"
	"code.linenisgreat.com/dodder/go/src/mike/store"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

type (
	envLocal     = env_local.Env
	envBox       = env_box.Env
	envWorkspace = env_workspace.Env
)

type Repo struct {
	envLocal
	envBox
	envWorkspace envWorkspace

	sunrise ids.Tai

	envRepo env_repo.Env
	config  store_config.StoreMutable

	storeAbbr    sku.AbbrStore
	dormantIndex dormant_index.Index

	storesInitialized bool
	typedBlobStore    typed_blob_store.Stores
	store             store.Store

	// TODO switch key to be workspace type
	workspaceStores map[ids.RepoId]*env_workspace.Store

	DormantCounter query.DormantCounter

	envLua env_lua.Env
}

func Make(
	env env_local.Env,
	options Options,
) *Repo {
	layoutOptions := env_repo.Options{
		BasePath: env.GetCLIConfig().BasePath,
	}

	var repoLayout env_repo.Env

	{
		var err error

		if repoLayout, err = env_repo.Make(
			env,
			layoutOptions,
		); err != nil {
			env.Cancel(err)
		}
	}

	return MakeWithLayout(options, repoLayout)
}

func MakeWithLayout(
	options Options,
	envRepo env_repo.Env,
) (repo *Repo) {
	repo = &Repo{
		config:         store_config.Make(),
		envLocal:       envRepo,
		envRepo:        envRepo,
		DormantCounter: query.MakeDormantCounter(),
	}

	repo.config.Reset()

	if err := repo.initialize(options); err != nil {
		repo.Cancel(err)
	}

	repo.After(errors.MakeFuncContextFromFuncErr(repo.Flush))

	return
}

// TODO investigate removing unnecessary resets like from organize
func (local *Repo) Reset() (err error) {
	return local.initialize(OptionsEmpty)
}

func (local *Repo) initialize(
	options Options,
) (err error) {
	if err = local.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// ui.Debug().Print(repo.layout.GetConfig().GetBlobStoreImmutableConfig().GetCompressionType())
	local.sunrise = ids.NowTai()

	if err = local.dormantIndex.Load(
		local.envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	objectFormat := object_inventory_format.FormatForVersion(
		local.envRepo.GetStoreVersion(),
	)

	boxFormatArchive := box_format.MakeBoxTransactedArchive(
		local.GetEnv(),
		local.GetConfig().PrintOptions.WithPrintTai(true),
	)

	if err = local.config.Initialize(
		local.envRepo,
		local.GetCLIConfig(),
	); err != nil {
		if options.GetAllowConfigReadError() {
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	if local.envWorkspace, err = env_workspace.Make(
		local.envRepo,
		local.config.GetConfigPtr(),
		local.PrinterFDDeleted(),
		local.GetEnvRepo(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if local.GetConfig().GetRepoType() != repo_type.TypeWorkingCopy {
		err = repo_type.ErrUnsupportedRepoType{
			Expected: repo_type.TypeWorkingCopy,
			Actual:   local.GetConfig().GetImmutableConfig().GetRepoType(),
		}

		return
	}

	if local.storeAbbr, err = store_abbr.NewIndexAbbr(
		local.config.GetConfig().PrintOptions,
		local.envRepo,
	); err != nil {
		err = errors.Wrapf(err, "failed to init abbr index")
		return
	}

	local.envBox = env_box.Make(
		local.envRepo,
		local.envWorkspace.GetStoreFS(),
		local.storeAbbr,
	)

	local.envLua = env_lua.Make(
		local.envRepo,
		local.GetStore(),
		local.SkuFormatBoxTransactedNoColor(),
	)

	// for _, rb := range u.GetConfig().Recipients {
	// 	if err = u.age.AddBech32PivYubikeyEC256(rb); err != nil {
	// 		errors.Wrap(err)
	// 		return
	// 	}
	// }

	local.typedBlobStore = typed_blob_store.MakeStores(
		local.envRepo,
		local.envLua,
		objectFormat,
		boxFormatArchive,
	)

	if err = local.store.Initialize(
		local.config,
		local.envRepo,
		local.envWorkspace,
		objectFormat,
		local.sunrise,
		local.envLua,
		local.makeQueryBuilder().
			WithDefaultGenres(ids.MakeGenre(genres.All()...)),
		boxFormatArchive,
		local.typedBlobStore,
		&local.dormantIndex,
		local.storeAbbr,
	); err != nil {
		err = errors.Wrapf(err, "failed to initialize store util")
		return
	}

	ui.Log().Printf(
		"store version: %s",
		local.GetConfig().GetImmutableConfig().GetStoreVersion(),
	)

	if err = local.envWorkspace.SetWorkspaceTypes(
		map[string]*env_workspace.Store{
			"browser": {
				StoreLike: store_browser.Make(
					local.config,
					local.GetEnvRepo(),
					local.PrinterTransactedDeleted(),
				),
			},
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = local.envWorkspace.SetSupplies(
		local.store.MakeSupplies(ids.RepoId{}),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Print("done initing checkout store")

	local.store.SetUIDelegate(local.GetUIStorePrinters())

	local.storesInitialized = true

	return
}

func (local *Repo) Flush() (err error) {
	waitGroup := errors.MakeWaitGroupParallel()

	if local.envWorkspace != nil {
		waitGroup.Do(local.envWorkspace.Flush)
	}

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) PrintMatchedDormantIfNecessary() {
	if !local.GetConfig().PrintOptions.PrintMatchedDormant {
		return
	}

	c := local.GetMatcherDormant().Count()
	ca := local.GetMatcherDormant().CountArchiviert()

	if c != 0 || ca == 0 {
		return
	}

	ui.Err().Printf("%d archived objects matched", c)
}

func (local *Repo) MakeObjectIdIndex() ids.Index {
	return ids.Index{}
}

func (local *Repo) GetMatcherDormant() query.DormantCounter {
	return local.DormantCounter
}

func (local *Repo) GetWorkspaceStoreForQuery(
	repoId ids.RepoId,
) (store_workspace.Store, bool) {
	return local.envWorkspace.GetStore(), true
}
