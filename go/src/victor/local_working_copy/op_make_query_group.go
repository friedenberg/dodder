package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/oscar/queries"
)

func (local *Repo) MakeExternalQueryGroup(
	metaBuilder queries.BuilderOption,
	externalQueryOptions sku.ExternalQueryOptions,
	args ...string,
) (queryGroup *queries.Query, err error) {
	builder := local.MakeQueryBuilderExcludingHidden(
		ids.MakeGenre(),
		metaBuilder,
	)

	if queryGroup, err = builder.BuildQueryGroupWithRepoId(
		externalQueryOptions,
		args...,
	); err != nil {
		err = errors.Wrap(err)
		return queryGroup, err
	}

	queryGroup.ExternalQueryOptions = externalQueryOptions

	return queryGroup, err
}

func (local *Repo) makeQueryBuilder() *queries.Builder {
	return queries.MakeBuilder(
		local.GetEnvRepo(),
		local.GetStore().GetTypedBlobStore(),
		local.GetStore().GetStreamIndex(),
		local.envLua.MakeLuaVMPoolBuilder(),
		local,
	)
}

func (local *Repo) MakeQueryBuilderExcludingHidden(
	genre ids.Genre,
	options queries.BuilderOption,
) *queries.Builder {
	if genre.IsEmpty() {
		genre = ids.MakeGenre(genres.Zettel)
	}

	options = queries.BuilderOptions(
		options,
		queries.BuilderOptionWorkspace(local),
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
	genress ids.Genre,
	options queries.BuilderOption,
) *queries.Builder {
	if genress.IsEmpty() {
		genress = ids.MakeGenre(genres.Zettel)
	}

	options = queries.BuilderOptions(
		options,
		queries.BuilderOptionWorkspace(local),
	)

	return local.makeQueryBuilder().
		WithDefaultGenres(genress).
		WithRepoId(ids.RepoId{}).
		WithFileExtensions(local.GetConfig().GetFileExtensions()).
		WithExpanders(local.GetStore().GetAbbrStore().GetAbbr()).
		WithOptions(options)
}
