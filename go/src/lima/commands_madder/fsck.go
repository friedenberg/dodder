package commands_madder

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/india/command_components_madder"
)

func init() {
	utility.AddCmd("fsck", &Fsck{})
}

type Fsck struct {
	command_components_madder.EnvBlobStore
	command_components_madder.BlobStore
}

// TODO add completion for blob store id's

func (cmd Fsck) Run(req command.Request) {
	envBlobStore := cmd.MakeEnvBlobStore(req)

	blobStores := cmd.MakeBlobStoresFromIndexesOrAll(req, envBlobStore)

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	for _, blobStore := range blobStores {
		ui.Out().Printf("%s", blobStore.NameWithIndex)
	}

	ui.Out().Print()

	for _, blobStore := range blobStores {
		ui.Out().Printf(
			"Verification for %s in progress...",
			blobStore.GetBlobStoreDescription(),
		)

		var count int
		var progressWriter env_ui.ProgressWriter

		countSuccessPtr := &count

		var blobErrors quiter.Slice[command_components_madder.BlobError]

		if err := errors.RunChildContextWithPrintTicker(
			envBlobStore,
			func(ctx interfaces.Context) {
				for digest, err := range blobStore.AllBlobs() {
					errors.ContextContinueOrPanic(ctx)
					// TODO keep track of blobs in a tridex and compare
					// subsequent stores

					if err != nil {
						blobErrors.Append(command_components_madder.BlobError{Err: err})

						continue
					}

					count++

					// TODO offer options:
					// - check existence
					// - verify can open
					// - print size
					// - compare against other blob stores
					if !blobStore.HasBlob(digest) {
						blobErrors.Append(
							command_components_madder.BlobError{Err: errors.Errorf("blob missing")},
						)

						continue
					}

					if err = blob_stores.VerifyBlob(
						ctx,
						blobStore,
						digest,
						&progressWriter,
					); err != nil {
						blobErrors.Append(
							command_components_madder.BlobError{Err: err},
						)

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
			envBlobStore.Cancel(err)
			return
		}

		ui.Out().Printf("blobs verified: %d", count)
		ui.Out().Printf(
			"blob bytes verified: %s",
			progressWriter.GetWrittenHumanString(),
		)

		command_components_madder.PrintBlobErrors(envBlobStore, blobErrors)
	}
}
