package interfaces

import (
	"hash"
	"io"
)

// TODO figure out if this is a blob id, digest, fingerprint, or what

type (
	BinaryId interface {
		GetBytes() []byte
	}

	MutableBinaryId interface {
		BinaryId
		SetBytes([]byte) error
		Reset()
	}

	BlobId interface {
		BinaryId
		BlobIdGetter
		GetSize() int
		GetType() string
		IsNull() bool
	}

	MutableBlobId interface {
		BlobId
		SetDigest(BlobId) error
		SetBytes([]byte) error
		Reset()
	}

	MutableGenericBlobId interface {
		MutableBlobId
		SetType(string) error
	}

	BlobIdGetter interface {
		GetBlobId() BlobId
	}

	EnvBlobId interface {
		GetType() string

		GetHash() (hash.Hash, FuncRepool)

		GetBlobId() MutableBlobId
		PutBlobId(BlobId)

		MakeWriteDigesterWithRepool() (WriteBlobIdGetter, FuncRepool)
		MakeWriteDigester() WriteBlobIdGetter
		MakeDigestFromHash(hash.Hash) (BlobId, error)
	}

	WriteBlobIdGetter interface {
		io.Writer
		BlobIdGetter
	}

	ReadBlobIdGetter interface {
		io.Reader
		BlobIdGetter
	}

	ReadCloseBlobIdGetter interface {
		io.WriterTo
		io.ReadCloser
		BlobIdGetter
	}

	WriteCloseBlobIdGetter interface {
		io.ReaderFrom
		io.WriteCloser
		BlobIdGetter
	}
)
