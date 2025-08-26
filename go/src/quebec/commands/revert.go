package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/flags"
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

func (c *Revert) SetFlagSet(f *flags.FlagSet) {
	c.LocalWorkingCopyWithQueryGroup.SetFlagSet(f)
	f.BoolVar(&c.Last, "last", false, "revert the last changes")
}

func (c Revert) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Zettel,
		genres.Tag,
		genres.Type,
		genres.InventoryList,
		genres.Repo,
	)
}

func (cmd Revert) Run(dep command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
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

	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Lock))

	switch {
	case cmd.Last:
		if err := cmd.runRevertFromLast(localWorkingCopy); err != nil {
			localWorkingCopy.Cancel(err)
		}

	default:
		if err := cmd.runRevertFromQuery(localWorkingCopy, queryGroup); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	localWorkingCopy.Must(errors.MakeFuncContextFromFuncErr(localWorkingCopy.Unlock))
}

func (c Revert) runRevertFromQuery(
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
				return
			}

			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (c Revert) runRevertFromLast(
	u *local_working_copy.Repo,
) (err error) {
	s := u.GetStore()

	var b *sku.Transacted

	if b, err = s.GetInventoryListStore().ReadLast(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for z, errIter := range s.GetInventoryListStore().IterInventoryList(b.GetBlobDigest()) {
		var cachedSku *sku.Transacted

		if cachedSku, err = u.GetStore().GetStreamIndex().ReadOneObjectIdTai(
			z.GetObjectId(),
			z.GetTai(),
		); errIter != nil {
			err = errors.Wrap(errIter)
			return
		}

		defer sku.GetTransactedPool().Put(cachedSku)

		rt := store.RevertId{
			ObjectId: cachedSku.GetObjectId(),
			Tai:      cachedSku.Metadata.Cache.ParentTai,
		}

		if err = u.GetStore().RevertTo(rt); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
