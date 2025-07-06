package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
)

func (local *Repo) MakeExternalQueryGroup(
	metaBuilder query.BuilderOption,
	externalQueryOptions sku.ExternalQueryOptions,
	args ...string,
) (queryGroup *query.Query, err error) {
	builder := local.MakeQueryBuilderExcludingHidden(ids.MakeGenre(), metaBuilder)

	if queryGroup, err = builder.BuildQueryGroupWithRepoId(
		externalQueryOptions,
		args...,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	queryGroup.ExternalQueryOptions = externalQueryOptions

	return
}

func (local *Repo) makeQueryBuilder() *query.Builder {
	return query.MakeBuilder(
		local.GetEnvRepo(),
		local.GetStore().GetTypedBlobStore(),
		local.GetStore().GetStreamIndex(),
		local.envLua.MakeLuaVMPoolBuilder(),
		local,
	)
}

func (local *Repo) MakeQueryBuilderExcludingHidden(
	genre ids.Genre,
	options query.BuilderOption,
) *query.Builder {
	if genre.IsEmpty() {
		genre = ids.MakeGenre(genres.Zettel)
	}

	envWorkspace := local.GetEnvWorkspace()

	options = query.BuilderOptions(
		options,
		query.BuilderOptionWorkspace{Env: envWorkspace},
	)

	return local.makeQueryBuilder().
		WithDefaultGenres(genre).
		WithRepoId(ids.RepoId{}).
		WithFileExtensions(local.GetConfig().GetFileExtensions()).
		WithExpanders(local.GetStore().GetAbbrStore().GetAbbr()).
		WithHidden(local.GetMatcherDormant()).
		WithOptions(options)
}

func (local *Repo) MakeQueryBuilder(
	dg ids.Genre,
	options query.BuilderOption,
) *query.Builder {
	if dg.IsEmpty() {
		dg = ids.MakeGenre(genres.Zettel)
	}

	envWorkspace := local.GetEnvWorkspace()

	options = query.BuilderOptions(
		options,
		query.BuilderOptionWorkspace{Env: envWorkspace},
	)

	return local.makeQueryBuilder().
		WithDefaultGenres(dg).
		WithRepoId(ids.RepoId{}).
		WithFileExtensions(local.GetConfig().GetFileExtensions()).
		WithExpanders(local.GetStore().GetAbbrStore().GetAbbr()).
		WithOptions(options)
}
