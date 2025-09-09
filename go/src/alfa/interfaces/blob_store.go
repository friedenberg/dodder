package interfaces

type (
	BlobIOWrapper interface {
		GetBlobEncryption() MarklId
		GetBlobCompression() IOWrapper
	}

	BlobIOWrapperGetter interface {
		GetBlobIOWrapper() BlobIOWrapper
	}

	BlobReaderFactory interface {
		MakeBlobReader(MarklId) (ReadCloseMarklIdGetter, error)
	}

	BlobWriterFactory interface {
		MakeBlobWriter(marklHashTypeId string) (WriteCloseMarklIdGetter, error)
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
