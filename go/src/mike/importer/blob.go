package importer

import (
	"time"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
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
	blobCopyResult := sku.BlobCopyResult{
		ObjectOrNil: object,
		CopyResult: blob_stores.CopyResult{
			BlobId: blobId,
		},
	}

	// when this is a dumb HTTP remote, we expect local to push the missing
	// objects to us after the import call

	blobCopyResult.SetBlobMissingLocally()

	if blobImporter.Src.HasBlob(blobId) {
		blobCopyResult.SetBlobExistsLocally()
	}

	if err = blobImporter.emitCopyResultIfNecessary(blobCopyResult); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (blobImporter *BlobImporter) emitCopyResultIfNecessary(
	copyResult sku.BlobCopyResult,
) (err error) {
	if blobImporter.CopierDelegate == nil {
		return
	}

	if err = blobImporter.CopierDelegate(copyResult); err != nil {
		err = errors.Wrap(err)
		return
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
			copyResult := sku.BlobCopyResult{
				ObjectOrNil: object,
			}

			blobImporter.Counts.Total++

			// TODO the written blob digest should be determined by an option
			copyResult.CopyResult = blob_stores.CopyBlobIfNecessary(
				blobImporter.EnvRepo,
				dst,
				blobImporter.Src,
				blobId,
				&progressWriter,
			)

			if copyResult.IsError() {
				blobImporter.Counts.Failed++
				ctx.Cancel(err)
			} else if copyResult.IsMissing() {
				blobImporter.Counts.Failed++
			} else if copyResult.Exists() {
				blobImporter.Counts.Ignored++
			} else {
				blobImporter.Counts.Succeeded++
			}

			if err = blobImporter.emitCopyResultIfNecessary(
				copyResult,
			); err != nil {
				err = errors.Wrap(err)
				return
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
