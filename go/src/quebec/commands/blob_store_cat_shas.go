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
	command.Register("blob_store-cat-shas", &BlobStoreCatShas{})
}

type BlobStoreCatShas struct {
	command_components.EnvRepo
}

func (cmd BlobStoreCatShas) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Blob,
	)
}

func (cmd BlobStoreCatShas) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)

	if err := envRepo.ReadAllShasForBlobs(
		func(sh *sha.Sha) (err error) {
			_, err = fmt.Fprintln(envRepo.GetUIFile(), sh)
			return
		},
	); err != nil {
		req.CancelWithError(err)
	}
}
