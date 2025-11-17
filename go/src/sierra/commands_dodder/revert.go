package commands_dodder

import (
	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/command"
	"code.linenisgreat.com/dodder/go/src/kilo/sku"
	"code.linenisgreat.com/dodder/go/src/mike/queries"
	"code.linenisgreat.com/dodder/go/src/romeo/store"
	"code.linenisgreat.com/dodder/go/src/sierra/command_components_dodder"
	"code.linenisgreat.com/dodder/go/src/sierra/local_working_copy"
)

func init() {
	utility.AddCmd("revert", &Revert{})
}

type Revert struct {
	command_components_dodder.LocalWorkingCopyWithQueryGroup

	Last bool
}

var _ interfaces.CommandComponentWriter = (*Revert)(nil)

func (md *Revert) SetFlagDefinitions(f interfaces.CLIFlagDefinitions) {
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
		queries.BuilderOptions(
			queries.BuilderOptionDefaultGenres(
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
	repo *local_working_copy.Repo,
	eq *queries.Query,
) (err error) {
	if err = repo.GetStore().QueryTransacted(
		eq,
		func(object *sku.Transacted) (err error) {
			revertId := store.RevertId{
				ObjectId: object.GetObjectId(),
				Sig:      object.GetMetadata().GetMotherObjectSig(),
			}

			if err = repo.GetStore().RevertTo(revertId); err != nil {
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
			Sig:      cachedSku.GetMetadata().GetMotherObjectSig(),
		}

		if err = repo.GetStore().RevertTo(rt); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}
