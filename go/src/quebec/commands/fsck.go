package commands

import (
	"flag"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/kilo/query"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"fsck",
		&Fsck{
			Genres: ids.MakeGenre(genres.Tag, genres.Type, genres.Zettel),
		},
	)
}

type Fsck struct {
	command_components.LocalWorkingCopyWithQueryGroup

	Genres ids.Genre
}

func (cmd *Fsck) SetFlagSet(f *flag.FlagSet) {
	cmd.LocalWorkingCopyWithQueryGroup.SetFlagSet(f)
	f.Var(&cmd.Genres, "genres", "")
}

func (cmd Fsck) Run(dep command.Request) {
	localWorkingCopy, queryGroup := cmd.MakeLocalWorkingCopyAndQueryGroup(
		dep,
		query.BuilderOptions(),
	)

	printer := localWorkingCopy.PrinterTransacted()

	if err := localWorkingCopy.GetStore().QueryTransacted(
		queryGroup,
		func(sk *sku.Transacted) (err error) {
			if !cmd.Genres.Contains(sk.GetGenre()) {
				return
			}

			blobSha := sk.GetBlobId()

			if localWorkingCopy.GetEnvRepo().GetDefaultBlobStore().HasBlob(blobSha) {
				return
			}

			if err = printer(sk); err != nil {
				err = errors.Wrap(err)
				return
			}

			return
		},
	); err != nil {
		localWorkingCopy.Cancel(err)
	}
}
