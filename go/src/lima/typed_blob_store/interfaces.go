package typed_blob_store

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	Format[T any, TPtr interfaces.Ptr[T]] interface {
		interfaces.SavedBlobFormatter
		interfaces.CoderReadWriter[TPtr]
	}
)

type TypedStore[
	A any,
	APtr interfaces.Ptr[A],
] interface {
	// TODO remove and replace with two-step process
	SaveBlobText(APtr) (interfaces.BlobId, int64, error)
	Format[A, APtr]
	// TODO remove
	interfaces.BlobPool[APtr]
}
