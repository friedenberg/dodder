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
	src interfaces.BlobStore,
	dsts ...interfaces.BlobStore,
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
	Src            interfaces.BlobStore
	Dsts           []interfaces.BlobStore

	Counts Counts
}

type Counts struct {
	Succeeded int
	Ignored   int
	Failed    int
	Total     int
}

func (blobImporter *BlobImporter) ImportBlobIfNecessary(
	blobId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	if len(blobImporter.Dsts) == 0 {
		return blobImporter.emitMissingBlob(blobId, object)
	}

	for _, blobStore := range blobImporter.Dsts {
		if err = blobImporter.importBlobIfNecessary(
			blobStore,
			blobId,
			object,
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobImporter *BlobImporter) emitMissingBlob(
	blobId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	// when this is a dumb HTTP remote, we expect local to push the missing
	// objects to us after the import call

	n := int64(-1)

	if blobImporter.Src.HasBlob(blobId) {
		n = -2
	}

	if blobImporter.CopierDelegate != nil {
		if err = blobImporter.CopierDelegate(
			sku.BlobCopyResult{
				Transacted: object,
				MarklId:    blobId,
				N:          n,
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (blobImporter *BlobImporter) importBlobIfNecessary(
	dst interfaces.BlobStore,
	blobId interfaces.MarklId,
	object *sku.Transacted,
) (err error) {
	var progressWriter env_ui.ProgressWriter

	if err = errors.RunChildContextWithPrintTicker(
		blobImporter.EnvRepo,
		func(ctx interfaces.Context) {
			var n int64

			blobImporter.Counts.Total++

			if n, err = blob_stores.CopyBlobIfNecessary(
				blobImporter.EnvRepo,
				dst,
				blobImporter.Src,
				blobId,
				&progressWriter,
			); err != nil {
				ui.Log().Print("copy failed", err, dst.GetBlobStoreDescription(), blobId)
				if errors.Is(err, env_dir.ErrBlobAlreadyExists{}) {
					blobImporter.Counts.Ignored++
					n = -3
					err = nil
				} else {
					blobImporter.Counts.Failed++
					// TODO add context that this could not be copied from the
					// remote blob
					// store
					err = errors.Wrap(err)
					return
				}
			} else {
				blobImporter.Counts.Succeeded++
			}

			if blobImporter.CopierDelegate != nil {
				if err = blobImporter.CopierDelegate(
					sku.BlobCopyResult{
						Transacted: object,
						MarklId:    blobId,
						N:          n,
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
