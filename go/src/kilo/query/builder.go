package query

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/lua"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/foxtrot/store_workspace"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/lima/typed_blob_store"
)

func MakeBuilder(
	envRepo env_repo.Env,
	typedBlobStore typed_blob_store.Stores,
	objectProbeIndex sku.ObjectProbeIndex,
	luaVMPoolBuilder *lua.VMPoolBuilder,
	workspaceStoreGetter store_workspace.StoreGetter,
) (b *Builder) {
	b = &Builder{
		envRepo:              envRepo,
		typedBlobStore:       typedBlobStore,
		objectProbeIndex:     objectProbeIndex,
		luaVMPoolBuilder:     luaVMPoolBuilder,
		workspaceStoreGetter: workspaceStoreGetter,
	}

	return
}

type Builder struct {
	envRepo                 env_repo.Env
	typedBlobStore          typed_blob_store.Stores
	objectProbeIndex        sku.ObjectProbeIndex
	luaVMPoolBuilder        *lua.VMPoolBuilder
	pinnedObjectIds         []pinnedObjectId
	pinnedExternalObjectIds []sku.ExternalObjectId
	workspaceStoreGetter    store_workspace.StoreGetter
	repoId                  ids.RepoId
	fileExtensions          interfaces.FileExtensions
	expanders               ids.Abbr
	hidden                  sku.Query
	doNotMatchEmpty         bool
	debug                   bool
	requireNonEmptyQuery    bool
	defaultQuery            string
	workspaceEnabled        bool

	options options
}

func (builder *Builder) makeState() *buildState {
	state := &buildState{
		options:      builder.options,
		builder:      builder,
		latentErrors: errors.MakeGroupBuilder(),
	}

	if builder.luaVMPoolBuilder != nil {
		state.luaVMPoolBuilder = builder.luaVMPoolBuilder.Clone()
	}

	state.group = state.makeGroup()

	state.pinnedObjectIds = make([]pinnedObjectId, len(builder.pinnedObjectIds))
	copy(state.pinnedObjectIds, builder.pinnedObjectIds)

	state.pinnedExternalObjectIds = make(
		[]sku.ExternalObjectId,
		len(builder.pinnedExternalObjectIds),
	)

	copy(state.pinnedExternalObjectIds, builder.pinnedExternalObjectIds)

	return state
}

func (builder *Builder) WithOptions(options BuilderOption) *Builder {
	if options == nil {
		return builder
	}

	applied := options.Apply(builder)

	if applied != nil {
		return applied
	}

	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithPermittedSigil(s ids.Sigil) *Builder {
	builder.options.permittedSigil.Add(s)
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithDoNotMatchEmpty() *Builder {
	builder.doNotMatchEmpty = true
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithRequireNonEmptyQuery() *Builder {
	builder.requireNonEmptyQuery = true
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithDebug() *Builder {
	builder.debug = true
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithRepoId(
	repoId ids.RepoId,
) *Builder {
	builder.repoId = repoId
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithFileExtensions(
	feg interfaces.FileExtensions,
) *Builder {
	builder.fileExtensions = feg
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithExpanders(
	expanders ids.Abbr,
) *Builder {
	builder.expanders = expanders
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithDefaultGenres(
	defaultGenres ids.Genre,
) *Builder {
	builder.options.defaultGenres = defaultGenres
	return builder
}

// TODO refactor into BuilderOption
func (builder *Builder) WithDefaultSigil(
	defaultSigil ids.Sigil,
) *Builder {
	builder.options.defaultSigil = defaultSigil
	return builder
}

func (builder *Builder) WithHidden(
	hidden sku.Query,
) *Builder {
	builder.hidden = hidden
	return builder
}

// TODO
func (builder *Builder) WithExternalLike(
	zts sku.SkuTypeSet,
) *Builder {
	for t := range zts.All() {
		if t.GetExternalObjectId().IsEmpty() {
			builder.pinnedObjectIds = append(
				builder.pinnedObjectIds,
				pinnedObjectId{
					Sigil: ids.SigilExternal,
					ObjectId: ObjectId{
						Exact:    true,
						ObjectId: t.GetObjectId(),
					},
				},
			)
		} else {
			if t.GetExternalObjectId().GetGenre() == genres.None {
				panic(
					errors.BadRequestf(
						"External object ID has an empty genre: %q",
						t.GetExternalObjectId(),
					),
				)
			}

			builder.pinnedExternalObjectIds = append(
				builder.pinnedExternalObjectIds,
				t.GetExternalObjectId(),
			)
		}
	}

	return builder
}

func (builder *Builder) WithTransacted(
	zts sku.TransactedSet,
	sigil ids.Sigil,
) *Builder {
	for t := range zts.All() {
		builder.pinnedObjectIds = append(
			builder.pinnedObjectIds,
			pinnedObjectId{
				Sigil: sigil,
				ObjectId: ObjectId{
					ObjectId: t.ObjectId.Clone(),
				},
			},
		)
	}

	return builder
}

func (builder *Builder) BuildQueryGroupWithRepoId(
	externalQueryOptions sku.ExternalQueryOptions,
	values ...string,
) (query *Query, err error) {
	state := builder.makeState()

	if builder.workspaceEnabled {
		ok := false

		state.workspaceStore, ok = builder.workspaceStoreGetter.GetWorkspaceStoreForQuery(
			externalQueryOptions.RepoId,
		)

		state.group.RepoId = externalQueryOptions.RepoId
		state.group.ExternalQueryOptions = externalQueryOptions

		if !ok {
			err = errors.ErrorWithStackf(
				"kasten not found: %q",
				externalQueryOptions.RepoId,
			)
			return
		}
	}

	if err = builder.build(state, values...); err != nil {
		err = errors.Wrap(err)
		return
	}

	query = state.group

	return
}

func (builder *Builder) BuildQueryGroup(
	args ...string,
) (group *Query, err error) {
	state := builder.makeState()

	if err = builder.build(state, args...); err != nil {
		err = errors.Wrap(err)
		return
	}

	group = state.group

	return
}

func (builder *Builder) build(state *buildState, values ...string) (err error) {
	var latent errors.GroupBuilder

	if err, latent = state.build(values...); err != nil {
		if !errors.IsBadRequest(err) {
			latent.Add(errors.Wrapf(err, "Query String: %q", values))
			err = latent.GetError()
		}

		errors.Wrap(err)

		return
	}

	if len(state.missingBlobs) > 0 {
		me := errors.MakeGroupBuilder()

		for _, e := range state.missingBlobs {
			me.Add(e)
		}

		err = me

		return
	}

	if builder.defaultQuery == "" {
		return
	}

	defaultQueryGroupState := state.copy()
	defaultQueryGroupState.options.defaultGenres = ids.MakeGenre(
		genres.All()...)

	if err, _ = defaultQueryGroupState.build(builder.defaultQuery); err != nil {
		err = errors.Wrap(err)
		return
	}

	state.group.defaultQuery = defaultQueryGroupState.group

	ui.Log().Print(state.group.StringDebug())

	return
}
