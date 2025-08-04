package importer

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/env_dir"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

func MakeBlobImporter(
	envRepo env_repo.Env,
	src env_repo.BlobStoreInitialized,
	dsts ...env_repo.BlobStoreInitialized,
) BlobImporter {
	return BlobImporter{
		EnvRepo: envRepo,
		Src:     src,
		Dsts:    dsts,
	}
}

type BlobImporter struct {
	EnvRepo        env_repo.Env
	CopierDelegate interfaces.FuncIter[sku.BlobCopyResult]
	Src            env_repo.BlobStoreInitialized
	Dsts           []env_repo.BlobStoreInitialized
	CountSuccess   int
	CountIgnored   int
	CountFailure   int
}

func (blobImporter *BlobImporter) ImportBlobIfNecessary(
	blobId interfaces.BlobId,
) (err error) {
	for _, blobStore := range blobImporter.Dsts {
		ui.Log().Print("copying", blobStore.GetBlobStoreDescription(), blobId)

		if err = blobImporter.importBlobIfNecessary(
			blobStore,
			blobId,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobImporter *BlobImporter) importBlobIfNecessary(
	dst interfaces.BlobStore,
	blobId interfaces.BlobId,
) (err error) {
	var progressWriter env_ui.ProgressWriter

	if err = errors.RunChildContextWithPrintTicker(
		blobImporter.EnvRepo,
		func(ctx interfaces.Context) {
			var n int64

			if n, err = blob_stores.CopyBlobIfNecessary(
				blobImporter.EnvRepo,
				dst,
				blobImporter.Src,
				blobId,
				&progressWriter,
			); err != nil {
				ui.Log().Print("copy failed", err, dst.GetBlobStoreDescription(), blobId)
				if errors.Is(err, env_dir.ErrBlobAlreadyExists{}) {
					blobImporter.CountIgnored++
					n = -3
					err = nil
				} else {
					blobImporter.CountFailure++
					// TODO add context that this could not be copied from the
					// remote blob
					// store
					err = errors.Wrap(err)
					return
				}
			} else {
				blobImporter.CountSuccess++
			}

			if blobImporter.CopierDelegate != nil {
				if err = blobImporter.CopierDelegate(
					sku.BlobCopyResult{
						BlobId: blobId,
						N:      n,
					},
				); err != nil {
					err = errors.Wrap(err)
					return
				}
			}
		},
		func(time time.Time) {
			ui.Err().Printf(
				"Copying %s... (%s written)",
				blobId,
				progressWriter.GetWrittenHumanString(),
			)
		},
		3*time.Second,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
