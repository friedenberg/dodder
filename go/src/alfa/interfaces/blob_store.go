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
		MakeBlobWriter(FormatHash) (BlobWriter, error)
	}

	BlobAccess interface {
		HasBlob(MarklId) bool
		BlobReaderFactory
		BlobWriterFactory
	}

	NamedBlobAccess interface {
		MakeNamedBlobReader(string) (BlobReader, error)
		MakeNamedBlobWriter(string) (BlobWriter, error)
	}

	BlobStore interface {
		BlobAccess
		BlobIOWrapperGetter

		GetBlobStoreDescription() string
		GetDefaultHashType() FormatHash
		AllBlobs() SeqError[MarklId]
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob(MarklId) (BLOB, FuncRepool, error)
	}
)
