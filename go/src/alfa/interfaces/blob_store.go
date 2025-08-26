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
		BlobReader(BlobId) (ReadCloseBlobIdGetter, error)
	}

	BlobWriter interface {
		BlobWriter() (WriteCloseBlobIdGetter, error)
	}

	Mover interface {
		io.WriteCloser
		io.ReaderFrom
		BlobIdGetter
	}

	BlobStoreConfig interface {
		GetBlobStoreType() string
	}

	BlobStore interface {
		GetBlobStoreConfig() BlobStoreConfig
		GetBlobStoreDescription() string
		HasBlob(BlobId) bool
		BlobReader
		BlobWriter
		AllBlobs() SeqError[BlobId]
		GetBlobIOWrapper() BlobIOWrapper
		Mover() (Mover, error)
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob2(BlobId) (BLOB, FuncRepool, error)
		// TODO replace with above
		GetBlob(BlobId) (BLOB, error)
		PutBlob(BLOB)
	}
)
