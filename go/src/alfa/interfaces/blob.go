package interfaces

import "iter"

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

	BlobStore interface {
		GetBlobStore() BlobStore
		HasBlob(sh Sha) (ok bool)
		BlobReader
		BlobWriter
	}

	LocalBlobStore interface {
		BlobStore
		GetLocalBlobStore() LocalBlobStore
		AllBlobs() iter.Seq2[Sha, error]
	}

	BlobStoreConfigImmutable interface {
		GetBlobStoreConfigImmutable() BlobStoreConfigImmutable
		GetBlobEncryption() BlobEncryption
		GetBlobCompression() BlobCompression
		GetLockInternalFiles() bool
	}

	// Blobs represent persisted files, like blobs in Git. Blobs are used by
	// Zettels, types, tags, config, and inventory lists.
	BlobPool[V any] interface {
		GetBlob(Sha) (V, error)
		PutBlob(V)
	}
)
