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

	BlobId        = MerkleId
	MutableBlobId = MutableMerkleId

	BlobIdGetter interface {
		GetBlobId() MerkleId
	}

	EnvBlobId interface {
		GetType() string

		GetHash() (hash.Hash, FuncRepool)

		// TODO rename
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
