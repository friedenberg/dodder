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
		BlobReader(Digest) (ReadCloserDigester, error)
	}

	BlobWriter interface {
		BlobWriter() (WriteCloserDigester, error)
	}

	Mover interface {
		io.WriteCloser
		io.ReaderFrom
		DigestGetter
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
	BlobPool[V any] interface {
		GetBlob(Digest) (V, error)
		PutBlob(V)
	}
)
