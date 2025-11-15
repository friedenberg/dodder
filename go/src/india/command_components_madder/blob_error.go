package command_components_madder

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/collections_slice"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_local"
)

type BlobError struct {
	BlobId interfaces.MarklId
	Err    error
}

func PrintBlobErrors(
	envLocal env_local.Env,
	blobErrors collections_slice.Slice[BlobError],
) {
	ui.Err().Printf("blobs with errors: %d", blobErrors.Len())

	bufferedWriter, repool := pool.GetBufferedWriter(envLocal.GetErr())
	defer repool()

	defer errors.ContextMustFlush(envLocal, bufferedWriter)

	for _, errorBlob := range blobErrors {
		if errorBlob.BlobId == nil {
			bufferedWriter.WriteString("(empty blob id): ")
		} else {
			fmt.Fprintf(bufferedWriter, "%s: ", errorBlob.BlobId)
		}

		ui.CLIErrorTreeEncoder.EncodeTo(errorBlob.Err, bufferedWriter)

		bufferedWriter.WriteByte('\n')
	}
}
