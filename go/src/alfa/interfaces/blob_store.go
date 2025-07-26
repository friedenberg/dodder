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
		BlobReader(BlobId) (ReadCloseDigester, error)
	}

	BlobWriter interface {
		BlobWriter() (WriteCloseDigester, error)
	}

	Mover interface {
		io.WriteCloser
		io.ReaderFrom
		BlobIdGetter
	}

	BlobStore interface {
		GetBlobStoreDescription() string
		HasBlob(sh BlobId) (ok bool)
		BlobReader
		BlobWriter
	}

	// TODO merge into BlobStore
	LocalBlobStore interface {
		BlobStore
		GetLocalBlobStore() LocalBlobStore
		GetBlobIOWrapper() BlobIOWrapper
		// TODO add context
		AllBlobs() SeqError[BlobId]
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
