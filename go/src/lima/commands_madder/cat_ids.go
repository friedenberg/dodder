package commands_madder

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/genres"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("cat-ids", &CatIds{})
}

type CatIds struct {
	command_components_madder.EnvBlobStore
	command_components_madder.BlobStore
}

func (cmd CatIds) CompletionGenres() ids.Genre {
	return ids.MakeGenre(
		genres.Blob,
	)
}

func (cmd CatIds) Run(req command.Request) {
	envBlobStore := cmd.MakeEnvBlobStore(req)

	blobStores := cmd.MakeBlobStoresFromIndexesOrAll(req, envBlobStore)

	var blobErrors quiter.Slice[command_components_madder.BlobError]

	for _, blobStore := range blobStores {
		cmd.runOne(envBlobStore, blobStore, &blobErrors)
	}

	command_components_madder.PrintBlobErrors(envBlobStore, blobErrors)
}

func (cmd CatIds) runOne(
	envBlobStore env_repo.BlobStoreEnv,
	blobStore blob_stores.BlobStoreInitialized,
	blobErrors *quiter.Slice[command_components_madder.BlobError],
) {
	for id, err := range blobStore.AllBlobs() {
		errors.ContextContinueOrPanic(envBlobStore)

		if err != nil {
			blobErrors.Append(
				command_components_madder.BlobError{BlobId: id, Err: err},
			)
		} else {
			envBlobStore.GetUI().Print(id)
		}
	}
}
