package interfaces

import "io"

type (
	SavedBlobFormatter interface {
		FormatSavedBlob(io.Writer, MarklId) (int64, error)
	}

	Format[T any] interface {
		SavedBlobFormatter
		CoderReadWriter[T]
	}

	TypedBlobStore[T any] interface {
		ParseTypedBlob(
			tipe ObjectId,
			blobSha MarklId) (common T, n int64, err error)

		PutTypedBlob(
			ObjectId,
			T,
		) error
	}
)
