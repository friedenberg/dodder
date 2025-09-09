package interfaces

import (
	"io"
)

type (
	BlobIOWrapper interface {
		GetBlobEncryption() MarklId
		GetBlobCompression() IOWrapper
	}

	BlobIOWrapperGetter interface {
		GetBlobIOWrapper() BlobIOWrapper
	}

	BlobReader interface {
		BlobReader(MarklId) (ReadCloseMarklIdGetter, error)
	}

	BlobWriter interface {
		BlobWriter(marklHashTypeId string) (WriteCloseMarklIdGetter, error)
	}

	Mover interface {
		io.WriteCloser
		io.ReaderFrom
		MarklIdGetter
	}

	BlobAccess interface {
		HasBlob(MarklId) bool
		BlobReader
		BlobWriter
		AllBlobs() SeqError[MarklId]
	}

	BlobStore interface {
		BlobAccess

		GetBlobStoreDescription() string
		GetDefaultHashType() HashType

		GetBlobIOWrapper() BlobIOWrapper
		// TODO rename to MakeMover
		Mover() (Mover, error)
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob(MarklId) (BLOB, FuncRepool, error)
	}
)
