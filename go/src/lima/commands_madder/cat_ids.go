package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	command.Register("blob_store-cat-ids", &CatIds{})
}

type CatIds struct {
	command_components_madder.EnvRepo
	command_components_madder.BlobStore
}

func (cmd CatIds) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Blob,
	)
}

func (cmd CatIds) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	var blobStoreIndexOrConfigPath string

	if req.RemainingArgCount() == 1 {
		blobStoreIndexOrConfigPath = req.PopArg(
			"blob store id or blob store config path",
		)
	}

	req.AssertNoMoreArgs()

	blobStore := cmd.MakeBlobStore(envRepo, blobStoreIndexOrConfigPath)

	for id, err := range blobStore.AllBlobs() {
		errors.ContextContinueOrPanic(envRepo)

		if err != nil {
			// ui.CLIErrorTreeEncoder.EncodeTo(err, envRepo.GetErr())
			envRepo.GetErr().Print(err)
		} else {
			envRepo.GetUI().Print(id)
		}
	}
}
