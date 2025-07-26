package interfaces

import (
	"hash"
	"io"
)

type (
	BlobId interface {
		BlobIdGetter
		GetBytes() []byte
		GetType() string
		IsNull() bool
	}

	MutableBlobId interface {
		BlobIdGetter
		SetBytes([]byte) error
		Reset()
	}

	BlobIdGetter interface {
		GetBlobId() BlobId
	}

	EnvBlobId interface {
		GetType() string

		GetHash() (hash.Hash, func())

		GetBlobId() BlobId
		PutBlobId(BlobId)

		MakeWriteDigesterWithRepool() (WriteBlobIdGetter, func())
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
