package commands

import (
	"flag"
	"io"
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/golf/command"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
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
}

func (cmd *BlobFsck) SetFlagSet(flagSet *flag.FlagSet) {
}

func (cmd BlobFsck) Run(req command.Request) {
	envRepo := cmd.MakeEnvRepo(req, false)
	blobStores := envRepo.GetBlobStores()

	// TODO output TAP
	ui.Out().Print("Blob Stores:")

	for i, blobStore := range blobStores {
		ui.Out().Printf("%d: %s", i, blobStore.GetBlobStoreDescription())
	}

	for _, blobStore := range blobStores {
		var countSuccess int
		var bytesVerified int64
		var progressWriter env_ui.ProgressWriter

		countSuccessPtr := &countSuccess

		type errorBlob struct {
			sha interfaces.Sha
			err error
		}

		var blobErrors []errorBlob

		if err := errors.RunChildContextWithPrintTicker(
			envRepo,
			func(ctx errors.Context) {
				ui.Out().Printf(
					"Verification for %s in progress...",
					blobStore.GetBlobStoreDescription(),
				)

				for sh, err := range blobStore.AllBlobs() {
					ctx.ContinueOrPanicOnDone()

					if err != nil {
						blobErrors = append(blobErrors, errorBlob{err: err})
						continue
					}

					if err = cmd.verifyBlob(ctx, blobStore, sh, &progressWriter); err != nil {
						blobErrors = append(blobErrors, errorBlob{err: err})
						continue
					}

					countSuccess++
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
			envRepo.CancelWithError(err)
			return
		}

		ui.Out().Printf("blobs verified: %d", countSuccess)
		ui.Out().Printf("blob bytes verified: %d", bytesVerified)
		ui.Out().Printf("blobs with errors: %d", len(blobErrors))

		for _, errorBlob := range blobErrors {
			ui.Out().Printf("%s: %s", errorBlob.sha, errorBlob.err)
		}
	}
}

func (cmd BlobFsck) verifyBlob(
	ctx errors.Context,
	blobStore interfaces.LocalBlobStore,
	sh interfaces.Sha,
	progressWriter io.Writer,
) (err error) {
	var readCloser interfaces.ShaReadCloser

	if readCloser, err = blobStore.BlobReader(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if _, err = io.Copy(progressWriter, readCloser); err != nil {
		err = errors.Wrap(err)
		return
	}

	expected := sha.Make(sh)

	if err = expected.AssertEqualsShaLike(readCloser.GetShaLike()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = readCloser.Close(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
