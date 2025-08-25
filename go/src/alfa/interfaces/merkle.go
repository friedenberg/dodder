package interfaces

import (
	"encoding"
	"hash"
	"io"
)

type (
	MerkleId interface {
		Stringer
		encoding.BinaryMarshaler
		GetBytes() []byte
		// TODO rethink size as it works completely different between sha and
		// merkle
		GetSize() int
		GetType() string
		IsNull() bool
	}

	MutableMerkleId interface {
		MerkleId
		Setter
		encoding.BinaryUnmarshaler
		SetMerkleId(tipe string, bites []byte) error
		Reset()
		ResetWithMerkleId(MerkleId)
	}

	BlobId interface {
		MerkleId
		BlobIdGetter
	}

	MutableBlobId interface {
		MutableMerkleId
		BlobId
		SetDigest(BlobId) error
	}

	// TODO design a better pattern for interfaces that have concrete
	// implementations and polymorphic implementations
	MutableGenericBlobId interface {
		MutableBlobId
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
