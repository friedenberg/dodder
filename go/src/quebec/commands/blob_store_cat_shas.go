package commands

import (
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
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
	var blobStore env_repo.BlobStoreInitialized

	if req.RemainingArgCount() == 0 {
		blobStore = envRepo.GetDefaultBlobStore()
	} else {
		blobStore = cmd.MakeBlobStore(envRepo, req.PopArg("blob store id or blob store config path"))
	}

	req.AssertNoMoreArgs()

	for sh, err := range blobStore.AllBlobs() {
		envRepo.ContinueOrPanicOnDone()

		if err != nil {
			envRepo.GetErr().Print(err)
		} else {
			envRepo.GetUI().Print(sh)
		}
	}
}
