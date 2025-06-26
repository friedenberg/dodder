package commands

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("cat-blob-shas", &CatBlobShas{})
}

type CatBlobShas struct {
	command_components.EnvRepo
}

func (c CatBlobShas) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Blob,
	)
}

func (c CatBlobShas) Run(dep command.Request) {
	repoLayout := c.MakeEnvRepo(dep, false)

	if err := repoLayout.ReadAllShasForBlobs(
		func(s *sha.Sha) (err error) {
			_, err = fmt.Fprintln(repoLayout.GetUIFile(), s)
			return
		},
	); err != nil {
		dep.CancelWithError(err)
	}
}
