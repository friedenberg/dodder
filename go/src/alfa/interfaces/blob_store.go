package interfaces

import "io"

type (
	BlobIOWrapper interface {
		GetBlobEncryption() MarklId
		GetBlobCompression() IOWrapper
	}

	BlobIOWrapperGetter interface {
		GetBlobIOWrapper() BlobIOWrapper
	}

	BlobReader interface {
		io.WriterTo
		io.ReadCloser
		MarklIdGetter
	}

	BlobWriter interface {
		io.ReaderFrom
		io.WriteCloser
		MarklIdGetter
	}

	BlobReaderFactory interface {
		MakeBlobReader(MarklId) (BlobReader, error)
	}

	BlobWriterFactory interface {
		MakeBlobWriter(marklHashTypeId string) (BlobWriter, error)
	}

	BlobAccess interface {
		HasBlob(MarklId) bool
		BlobReaderFactory
		BlobWriterFactory
		AllBlobs() SeqError[MarklId]
	}

	BlobStore interface {
		BlobAccess
		BlobIOWrapperGetter

		GetBlobStoreDescription() string
		GetDefaultHashType() HashType
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob(MarklId) (BLOB, FuncRepool, error)
	}
)
