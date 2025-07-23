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
		BlobReader(Digest) (ReadCloseDigester, error)
	}

	BlobWriter interface {
		BlobWriter() (WriteCloseDigester, error)
	}

	Mover interface {
		io.WriteCloser
		io.ReaderFrom
		Digester
	}

	BlobStore interface {
		GetBlobStoreDescription() string
		HasBlob(sh Digest) (ok bool)
		BlobReader
		BlobWriter
	}

	// TODO merge into BlobStore
	LocalBlobStore interface {
		BlobStore
		GetLocalBlobStore() LocalBlobStore
		GetBlobIOWrapper() BlobIOWrapper
		// TODO add context
		AllBlobs() SeqError[Digest]
		Mover() (Mover, error)
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[BLOB any] interface {
		GetBlob2(Digest) (BLOB, FuncRepool, error)
		// TODO replace with above
		GetBlob(Digest) (BLOB, error)
		PutBlob(BLOB)
	}
)
