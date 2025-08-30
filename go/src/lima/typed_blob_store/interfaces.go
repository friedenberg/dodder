package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO replace with coders

type (
	Format[T any, TPtr interfaces.Ptr[T]] interface {
		interfaces.SavedBlobFormatter
		interfaces.CoderReadWriter[TPtr]
	}
)

type TypedStore[
	BLOB any,
	BLOB_PTR interfaces.Ptr[BLOB],
] interface {
	// TODO remove and replace with two-step process
	SaveBlobText(BLOB_PTR) (interfaces.MarklId, int64, error)
	Format[BLOB, BLOB_PTR]
	// TODO remove
	interfaces.BlobPool[BLOB_PTR]
}
