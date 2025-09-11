package sku

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/echo/fd"
	"code.linenisgreat.com/dodder/go/src/hotel/blob_stores"
)

type BlobCopyResult struct {
	ObjectOrNil *Transacted // may be nil
	blob_stores.CopyResult
}

func (copyResult BlobCopyResult) String() string {
	var stringsBuilder strings.Builder

	stringsBuilder.WriteString("BlobCopyResult{")

	if copyResult.ObjectOrNil == nil {
		stringsBuilder.WriteString("ObjectId: nil, ")
	} else {
		fmt.Fprintf(&stringsBuilder, "ObjectId: %s, ", copyResult.ObjectOrNil.GetObjectId())
	}

	stringsBuilder.WriteString(copyResult.CopyResult.String())

	stringsBuilder.WriteString("}")

	return stringsBuilder.String()
}

func MakeBlobCopierDelegate(
	fd fd.Std,
	includeAlreadyExists bool,
) func(BlobCopyResult) error {
	if !includeAlreadyExists {
		return func(result BlobCopyResult) error {
			if result.BlobId == nil {
				err := errors.Errorf("empty blob id")
				return err
			}

			bytesWritten, state := result.GetBytesWrittenAndState()

			switch state {
			case blob_stores.CopyResultStateExistsLocallyAndRemotely,
				blob_stores.CopyResultStateExistsLocally:
				return nil

			default:
				return fd.Printf(
					"copied Blob %s (%s)",
					result.BlobId,
					ui.GetHumanBytesString(uint64(bytesWritten)),
				)
			}
		}
	} else {
		return func(result BlobCopyResult) error {
			if result.BlobId == nil {
				err := errors.Errorf("empty blob id")
				return err
			}

			bytesWritten, state := result.GetBytesWrittenAndState()

			switch state {
			case blob_stores.CopyResultStateExistsLocallyAndRemotely, blob_stores.CopyResultStateExistsLocally:
				return fd.Printf(
					"Blob %s already exists",
					result.BlobId,
				)

			default:
				return fd.Printf(
					"copied Blob %s (%s)",
					result.BlobId,
					ui.GetHumanBytesString(uint64(bytesWritten)),
				)
			}
		}
	}
}
