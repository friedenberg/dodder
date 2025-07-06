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
			env.CancelWithError(err)
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
		repo.CancelWithError(err)
	}

	repo.After(repo.Flush)

	return
}

// TODO investigate removing unnecessary resets like from organize
func (u *Repo) Reset() (err error) {
	return u.initialize(OptionsEmpty)
}

func (repo *Repo) initialize(
	options Options,
) (err error) {
	if err = repo.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	// ui.Debug().Print(repo.layout.GetConfig().GetBlobStoreImmutableConfig().GetCompressionType())
	repo.sunrise = ids.NowTai()

	if err = repo.dormantIndex.Load(
		repo.envRepo,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	objectFormat := object_inventory_format.FormatForVersion(
		repo.envRepo.GetStoreVersion(),
	)

	boxFormatArchive := box_format.MakeBoxTransactedArchive(
		repo.GetEnv(),
		repo.GetConfig().GetCLIConfig().PrintOptions.WithPrintTai(true),
	)

	if err = repo.config.Initialize(
		repo.envRepo,
		repo.GetCLIConfig(),
	); err != nil {
		if options.GetAllowConfigReadError() {
			err = nil
		} else {
			err = errors.Wrapf(
				err,
				"CompressionType: %q",
				repo.envRepo.GetConfigPrivate().Blob.GetBlobStoreConfigImmutable().GetBlobCompression(),
			)
			return
		}
	}

	if repo.envWorkspace, err = env_workspace.Make(
		repo.envRepo,
		repo.config,
		repo.PrinterFDDeleted(),
		repo.GetEnvRepo(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if repo.GetConfig().GetRepoType() != repo_type.TypeWorkingCopy {
		err = repo_type.ErrUnsupportedRepoType{
			Expected: repo_type.TypeWorkingCopy,
			Actual:   repo.GetConfig().GetImmutableConfig().GetRepoType(),
		}

		return
	}

	if repo.storeAbbr, err = store_abbr.NewIndexAbbr(
		repo.config.GetCLIConfig().PrintOptions,
		repo.envRepo,
	); err != nil {
		err = errors.Wrapf(err, "failed to init abbr index")
		return
	}

	repo.envBox = env_box.Make(
		repo.envRepo,
		repo.envWorkspace.GetStoreFS(),
		repo.storeAbbr,
	)

	repo.envLua = env_lua.Make(
		repo.envRepo,
		repo.GetStore(),
		repo.SkuFormatBoxTransactedNoColor(),
	)

	// for _, rb := range u.GetConfig().Recipients {
	// 	if err = u.age.AddBech32PivYubikeyEC256(rb); err != nil {
	// 		errors.Wrap(err)
	// 		return
	// 	}
	// }

	repo.typedBlobStore = typed_blob_store.MakeStores(
		repo.envRepo,
		repo.envLua,
		objectFormat,
		boxFormatArchive,
	)

	if err = repo.store.Initialize(
		repo.config,
		repo.envRepo,
		repo.envWorkspace,
		objectFormat,
		repo.sunrise,
		repo.envLua,
		repo.makeQueryBuilder().
			WithDefaultGenres(ids.MakeGenre(genres.All()...)),
		boxFormatArchive,
		repo.typedBlobStore,
		&repo.dormantIndex,
		repo.storeAbbr,
	); err != nil {
		err = errors.Wrapf(err, "failed to initialize store util")
		return
	}

	ui.Log().Printf(
		"store version: %s",
		repo.GetConfig().GetImmutableConfig().GetStoreVersion(),
	)

	if err = repo.envWorkspace.SetWorkspaceTypes(
		map[string]*env_workspace.Store{
			"browser": {
				StoreLike: store_browser.Make(
					repo.config,
					repo.GetEnvRepo(),
					repo.PrinterTransactedDeleted(),
				),
			},
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = repo.envWorkspace.SetSupplies(
		repo.store.MakeSupplies(ids.RepoId{}),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	ui.Log().Print("done initing checkout store")

	repo.store.SetUIDelegate(repo.GetUIStorePrinters())

	repo.storesInitialized = true

	return
}

func (repo *Repo) Flush() (err error) {
	waitGroup := errors.MakeWaitGroupParallel()

	if repo.envWorkspace != nil {
		waitGroup.Do(repo.envWorkspace.Flush)
	}

	if err = waitGroup.GetError(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (u *Repo) PrintMatchedDormantIfNecessary() {
	if !u.GetConfig().GetCLIConfig().PrintOptions.PrintMatchedDormant {
		return
	}

	c := u.GetMatcherDormant().Count()
	ca := u.GetMatcherDormant().CountArchiviert()

	if c != 0 || ca == 0 {
		return
	}

	ui.Err().Printf("%d archived objects matched", c)
}

func (u *Repo) MakeObjectIdIndex() ids.Index {
	return ids.Index{}
}

func (u *Repo) GetMatcherDormant() query.DormantCounter {
	return u.DormantCounter
}

func (repo *Repo) GetWorkspaceStoreForQuery(
	repoId ids.RepoId,
) (store_workspace.Store, bool) {
	return repo.envWorkspace.GetStore(), true
}
