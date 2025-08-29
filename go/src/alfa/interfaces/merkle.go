package interfaces

import (
	"encoding"
	"hash"
	"io"
)

type (
	BlobId interface {
		Stringer
		encoding.BinaryMarshaler
		encoding.TextMarshaler
		// io.WriterTo
		GetBytes() []byte
		// TODO rethink size as it works completely different between sha and
		// merkle
		GetSize() int
		GetType() string
		IsNull() bool
	}

	MutableBlobId interface {
		BlobId
		Setter
		encoding.BinaryUnmarshaler
		encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMerkleId(tipe string, bites []byte) error
		Reset()
		ResetWithMerkleId(BlobId)
	}

	BlobIdGetter interface {
		GetBlobId() BlobId
	}

	Hash interface {
		hash.Hash
		GetType() string
		GetBlobId() (MutableBlobId, FuncRepool)
	}

	HashType interface {
		GetBlobIdForString(input string) (BlobId, FuncRepool)
		// TODO rename
		FromStringFormat(format string, args ...any) (BlobId, FuncRepool)
	}

	EnvBlobId interface {
		GetType() string

		GetHash() (hash.Hash, FuncRepool)

		GetBlobId() MutableBlobId
		PutBlobId(BlobId)
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
