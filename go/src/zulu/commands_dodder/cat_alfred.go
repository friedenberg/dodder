package commands_dodder

import (
	"bufio"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/echo/genres"
	"code.linenisgreat.com/dodder/go/src/foxtrot/alfred"
	"code.linenisgreat.com/dodder/go/src/foxtrot/ids"
	"code.linenisgreat.com/dodder/go/src/kilo/command"
	"code.linenisgreat.com/dodder/go/src/lima/sku"
	"code.linenisgreat.com/dodder/go/src/mike/alfred_sku"
	"code.linenisgreat.com/dodder/go/src/papa/queries"
	"code.linenisgreat.com/dodder/go/src/yankee/command_components_dodder"
)

func init() {
	utility.AddCmd("cat-alfred", &CatAlfred{})
}

type CatAlfred struct {
	command_components_dodder.LocalWorkingCopyWithQueryGroup

	genres.Genre
}

var _ interfaces.CommandComponentWriter = (*CatAlfred)(nil)

func (cmd *CatAlfred) SetFlagDefinitions(
	flagDefinitions interfaces.CLIFlagDefinitions,
) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagDefinitions(flagDefinitions)
	flagDefinitions.Var(
		&cmd.Genre,
		"genre",
		"extract this element from all matching objects",
	)
}

func (c CatAlfred) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Tag,
		genres.Type,
		genres.Zettel,
	)
}

func (cmd CatAlfred) Run(dep command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		dep,
		queries.BuilderOptions(
			queries.BuilderOptionDefaultGenres(
				genres.Tag,
				genres.Type,
				genres.Zettel,
			),
		),
	)

	// this command does its own error handling
	wo := bufio.NewWriter(localWorkingCopy.GetUIFile())
	defer errors.ContextMustFlush(localWorkingCopy, wo)

	var aiw alfred.Writer

	itemPool := alfred.MakeItemPool()

	switch cmd.Genre {
	case genres.Type, genres.Tag:
		{
			var err error

			if aiw, err = alfred.NewDebouncingWriter(localWorkingCopy.GetUIFile()); err != nil {
				localWorkingCopy.Cancel(err)
			}
		}

	default:
		{
			var err error

			if aiw, err = alfred.NewWriter(localWorkingCopy.GetUIFile(), itemPool); err != nil {
				localWorkingCopy.Cancel(err)
			}
		}
	}

	var aw *alfred_sku.Writer

	{
		var err error

		if aw, err = alfred_sku.New(
			wo,
			localWorkingCopy.GetStore().GetAbbrStore().GetAbbr(),
			localWorkingCopy.SkuFormatBoxTransactedNoColor(),
			aiw,
			itemPool,
		); err != nil {
			localWorkingCopy.Cancel(err)
		}
	}

	defer errors.ContextMustClose(localWorkingCopy, aw)

	if err := localWorkingCopy.GetStore().QueryTransacted(
		queryGroup,
		func(object *sku.Transacted) (err error) {
			switch cmd.Genre {
			case genres.Tag:
				for tag := range object.GetMetadata().GetTags().All() {
					var tagObject *sku.Transacted

					if tagObject, err = localWorkingCopy.GetStore().ReadTransactedFromObjectId(
						tag,
					); err != nil {
						if collections.IsErrNotFound(err) {
							err = nil
							tagObject = sku.GetTransactedPool().Get()
							defer sku.GetTransactedPool().Put(tagObject)
							tagObject.ObjectId.ResetWithIdLike(tag)
						} else {
							err = errors.Wrap(err)
							return err
						}
					}

					if err = aw.PrintOne(tagObject); err != nil {
						err = errors.Wrap(err)
						return err
					}
				}

			case genres.Type:
				tipe := object.GetType()

				if tipe.GetType().IsEmpty() {
					return err
				}

				if object, err = localWorkingCopy.GetStore().ReadTransactedFromObjectId(
					tipe.GetType(),
				); err != nil {
					err = errors.Wrap(err)
					return err
				}

				if err = aw.PrintOne(object); err != nil {
					err = errors.Wrap(err)
					return err
				}

			default:
				if err = aw.PrintOne(object); err != nil {
					err = errors.Wrap(err)
					return err
				}
			}

			return err
		},
	); err != nil {
		aw.WriteError(err)
		err = nil
		return
	}
}
