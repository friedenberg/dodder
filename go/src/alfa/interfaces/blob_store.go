package interfaces

import (
	"io"
)

type (
	BlobCompression interface {
		CommandLineIOWrapper
		GetBlobCompression() BlobCompression
	}

	BlobEncryption interface {
		CommandLineIOWrapper
		GetBlobEncryption() BlobEncryption
	}

	BlobIOWrapper interface {
		GetBlobEncryption() BlobEncryption
		GetBlobCompression() BlobCompression
	}

	BlobIOWrapperGetter interface {
		GetBlobIOWrapper() BlobIOWrapper
	}

	BlobReader interface {
		BlobReader(MarklId) (ReadCloseMarklIdGetter, error)
	}

	BlobWriter interface {
		BlobWriter() (WriteCloseMarklIdGetter, error)
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
