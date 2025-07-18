package commands

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register("blob_store-cat-shas", &BlobStoreCatShas{})
}

type BlobStoreCatShas struct {
	command_components.EnvRepo
	command_components.BlobStore
}

func (cmd BlobStoreCatShas) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Blob,
	)
}

func (cmd BlobStoreCatShas) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	var blobStoreIndexOrConfigPath string

	if req.RemainingArgCount() == 1 {
		blobStoreIndexOrConfigPath = req.PopArg(
			"blob store id or blob store config path",
		)
	}

	req.AssertNoMoreArgs()

	blobStore := cmd.MakeBlobStore(envRepo, blobStoreIndexOrConfigPath)

	for sh, err := range blobStore.AllBlobs() {
		errors.ContextContinueOrPanic(envRepo)

		if err != nil {
			envRepo.GetErr().Print(err)
		} else {
			envRepo.GetUI().Print(sh)
		}
	}
}
