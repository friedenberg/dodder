package commands

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/papa/command_components"
)

func init() {
	command.Register(
		"blob_store-fsck",
		&BlobFsck{},
	)
}

type BlobFsck struct {
	command_components.EnvRepo
	command_components.BlobStore
}

// TODO add completion for blob store id's

func (cmd BlobFsck) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)

	blobStoreIds := req.PopArgs()
	blobStores := envRepo.GetBlobStores()

	if len(blobStoreIds) > 0 {
		blobStores = blobStores[:0]
		for _, id := range blobStoreIds {
			blobStores = append(blobStores, cmd.MakeBlobStore(envRepo, id))
		}
	}

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	for i, blobStore := range blobStores {
		ui.Out().Printf("%d: %s", i, blobStore.Name)
	}

	for _, blobStore := range blobStores {
		ui.Out().Printf(
			"Verification for %s in progress...",
			blobStore.GetBlobStoreDescription(),
		)

		var count int
		var bytesVerified int64
		var progressWriter env_ui.ProgressWriter

		countSuccessPtr := &count

		type errorBlob struct {
			sha interfaces.BlobId
			err error
		}

		var blobErrors []errorBlob

		if err := errors.RunChildContextWithPrintTicker(
			envRepo,
			func(ctx interfaces.Context) {
				for digest, err := range blobStore.AllBlobs() {
					errors.ContextContinueOrPanic(ctx)
					// TODO keep track of blobs in a tridex and compare
					// subsequent stores

					if err != nil {
						blobErrors = append(blobErrors, errorBlob{err: err})
						continue
					}

					count++

					// TODO offer options:
					// - check existence
					// - verify can open
					// - print size
					// - compare against other blob stores
					if !blobStore.HasBlob(digest) {
						blobErrors = append(blobErrors, errorBlob{err: errors.Errorf("blob missing")})
						continue
					}

					if err = blob_stores.VerifyBlob(
						ctx,
						blobStore,
						digest,
						&progressWriter,
					); err != nil {
						blobErrors = append(blobErrors, errorBlob{err: err})
						continue
					}
				}
			},
			func(time time.Time) {
				ui.Out().Printf(
					"%d blobs / %s verified, %d errors",
					*countSuccessPtr,
					progressWriter.GetWrittenHumanString(),
					len(blobErrors),
				)
			},
			3*time.Second,
		); err != nil {
			envRepo.Cancel(err)
			return
		}

		ui.Out().Printf("blobs verified: %d", count)
		ui.Out().Printf("blob bytes verified: %d", bytesVerified)
		ui.Out().Printf("blobs with errors: %d", len(blobErrors))

		for _, errorBlob := range blobErrors {
			ui.Out().Printf("%s: %s", errorBlob.sha, errorBlob.err)
		}
	}
}
