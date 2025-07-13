package interfaces

import (
	"io"
	"iter"
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

	BlobReader interface {
		BlobReader(Sha) (ShaReadCloser, error)
	}

	BlobWriter interface {
		BlobWriter() (ShaWriteCloser, error)
	}

	Mover interface {
		io.WriteCloser
		GetShaLike() Sha
	}

	BlobStore interface {
		GetBlobStore() BlobStore
		HasBlob(sh Sha) (ok bool)
		BlobReader
		BlobWriter
	}

	// TODO merge into BlobStore
	LocalBlobStore interface {
		BlobStore
		GetLocalBlobStore() LocalBlobStore
		AllBlobs() iter.Seq2[Sha, error]
		Mover() (Mover, error)
	}

	BlobStoreIOWrapper interface {
		GetBlobEncryption() BlobEncryption
		GetBlobCompression() BlobCompression
	}

	BlobStoreConfigImmutable interface {
		GetBlobStoreConfigImmutable() BlobStoreConfigImmutable
		BlobStoreIOWrapper
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[V any] interface {
		GetBlob(Sha) (V, error)
		PutBlob(V)
	}
)
