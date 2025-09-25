package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/mike/store"
	"code.linenisgreat.com/dodder/go/src/november/local_working_copy"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("revert", &Revert{})
}

type Revert struct {
	command_components.LocalWorkingCopyWithQueryGroup

	Last bool
}

var _ interfaces.CommandComponentWriter = (*Revert)(nil)

func (md *Revert) SetFlagDefinitions(f interfaces.CommandLineFlagDefinitions) {
	md.LocalWorkingCopyWithQueryGroup.SetFlagDefinitions(f)
	f.BoolVar(&md.Last, "last", false, "revert the last changes")
}

func (md Revert) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Zettel,
		genres.Tag,
		genres.Type,
		genres.InventoryList,
		genres.Repo,
	)
}

func (md Revert) Run(dep command.Request) {
	localWorkingCopy, queryGroup := md.MakeLocalWorkingCopyAndQueryGroup(
		dep,
		query.BuilderOptions(
			query.BuilderOptionDefaultGenres(
				genres.Zettel,
				genres.Tag,
				genres.Type,
				genres.Repo,
			),
		),
	)

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock),
	)

	switch {
	case md.Last:
		if err := md.runRevertFromLast(localWorkingCopy); err != nil {
			localWorkingCopy.Cancel(err)
		}

	default:
		if err := md.runRevertFromQuery(localWorkingCopy, queryGroup); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	localWorkingCopy.Must(
		errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock),
	)
}

func (md Revert) runRevertFromQuery(
	u *local_working_copy.Repo,
	eq *query.Query,
) (err error) {
	if err = u.GetStore().QueryTransacted(
		eq,
		func(z *sku.Transacted) (err error) {
			rt := store.RevertId{
				ObjectId: z.GetObjectId(),
				Tai:      z.Metadata.Cache.ParentTai,
			}

			if err = u.GetStore().RevertTo(rt); err != nil {
				err = errors.Wrap(err)
				return err
			}

			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (md Revert) runRevertFromLast(
	repo *local_working_copy.Repo,
) (err error) {
	stoar := repo.GetStore()

	var listObject *sku.Transacted

	if listObject, err = stoar.GetInventoryListStore().ReadLast(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	seq := stoar.GetInventoryListStore().AllInventoryListContents(
		listObject.GetBlobDigest(),
	)

	for object, errIter := range seq {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return err
		}

		var cachedSku *sku.Transacted

		if cachedSku, err = repo.GetStore().GetStreamIndex().ReadOneObjectIdTai(
			object.GetObjectId(),
			object.GetTai(),
		); err != nil {
			err = errors.Wrap(errIter)
			return err
		}

		defer sku.GetTransactedPool().Put(cachedSku)

		rt := store.RevertId{
			ObjectId: cachedSku.GetObjectId(),
			Tai:      cachedSku.Metadata.Cache.ParentTai,
		}

		if err = repo.GetStore().RevertTo(rt); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
