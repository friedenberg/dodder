package interfaces

import "io"

type (
	SavedBlobFormatter interface {
		FormatSavedBlob(io.Writer, Sha) (int64, error)
	}

	Format[T any] interface {
		SavedBlobFormatter
		CoderReadWriter[T]
	}

	TypedBlobStore[T any] interface {
		ParseTypedBlob(
			tipe ObjectId,
			blobSha Sha,
		) (common T, n int64, err error)

		PutTypedBlob(
			ObjectId,
			T,
		) error
	}
)
