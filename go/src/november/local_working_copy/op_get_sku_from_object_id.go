package local_working_copy

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	pkg_query "code.linenisgreat.com/dodder/go/src/kilo/query"
)

// TODO rename
func (local *Repo) GetZettelFromObjectId(
	objectIdString string,
) (sk *sku.Transacted, err error) {
	builder := local.MakeQueryBuilder(ids.MakeGenre(genres.Zettel), nil)

	var query *pkg_query.Query

	if query, err = builder.BuildQueryGroupWithRepoId(
		sku.ExternalQueryOptions{},
		objectIdString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sk, err = local.GetStore().QueryExactlyOneExternal(query); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (local *Repo) GetObjectFromObjectId(
	objectIdString string,
) (sk *sku.Transacted, err error) {
	builder := local.MakeQueryBuilder(ids.MakeGenre(genres.All()...), nil)

	var queryGroup *pkg_query.Query

	if queryGroup, err = builder.BuildQueryGroupWithRepoId(
		sku.ExternalQueryOptions{},
		objectIdString,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if sk, err = local.GetStore().QueryExactlyOneExternal(queryGroup); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
