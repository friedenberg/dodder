package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
)

func (repo *Repo) MakeExternalQueryGroup(
	metaBuilder query.BuilderOption,
	externalQueryOptions sku.ExternalQueryOptions,
	args ...string,
) (queryGroup *query.Query, err error) {
	builder := repo.MakeQueryBuilderExcludingHidden(ids.MakeGenre(), metaBuilder)

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

func (repo *Repo) makeQueryBuilder() *query.Builder {
	return query.MakeBuilder(
		repo.GetEnvRepo(),
		repo.GetStore().GetTypedBlobStore(),
		repo.GetStore().GetStreamIndex(),
		repo.envLua.MakeLuaVMPoolBuilder(),
		repo,
	)
}

func (repo *Repo) MakeQueryBuilderExcludingHidden(
	genre ids.Genre,
	options query.BuilderOption,
) *query.Builder {
	if genre.IsEmpty() {
		genre = ids.MakeGenre(genres.Zettel)
	}

	envWorkspace := repo.GetEnvWorkspace()

	options = query.BuilderOptions(
		options,
		query.BuilderOptionWorkspace{Env: envWorkspace},
	)

	return repo.makeQueryBuilder().
		WithDefaultGenres(genre).
		WithRepoId(ids.RepoId{}).
		WithFileExtensions(repo.GetConfig().GetFileExtensions()).
		WithExpanders(repo.GetStore().GetAbbrStore().GetAbbr()).
		WithHidden(repo.GetMatcherDormant()).
		WithOptions(options)
}

func (repo *Repo) MakeQueryBuilder(
	dg ids.Genre,
	options query.BuilderOption,
) *query.Builder {
	if dg.IsEmpty() {
		dg = ids.MakeGenre(genres.Zettel)
	}

	envWorkspace := repo.GetEnvWorkspace()

	options = query.BuilderOptions(
		options,
		query.BuilderOptionWorkspace{Env: envWorkspace},
	)

	return repo.makeQueryBuilder().
		WithDefaultGenres(dg).
		WithRepoId(ids.RepoId{}).
		WithFileExtensions(repo.GetConfig().GetFileExtensions()).
		WithExpanders(repo.GetStore().GetAbbrStore().GetAbbr()).
		WithOptions(options)
}
