package interfaces

import (
	"io"
)

type (
	BlobIOWrapper interface {
		GetBlobEncryption() CommandLineIOWrapper
		GetBlobCompression() CommandLineIOWrapper
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

	BlobStoreConfig interface {
		GetBlobStoreType() string
	}

	BlobStore interface {
		GetBlobStoreConfig() BlobStoreConfig
		GetBlobStoreDescription() string
		HasBlob(MarklId) bool
		BlobReader
		BlobWriter
		AllBlobs() SeqError[MarklId]
		GetBlobIOWrapper() BlobIOWrapper
		Mover() (Mover, error)
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob(MarklId) (BLOB, FuncRepool, error)
	}
)
