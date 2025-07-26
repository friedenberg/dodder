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

		// TODO rename to "MakeDigest"
		GetBlobId() BlobId
		PutBlobId(BlobId)

		MakeWriteDigesterWithRepool() (WriteDigester, func())
		MakeWriteDigester() WriteDigester
		MakeDigestFromHash(hash.Hash) (BlobId, error)
	}

	WriteDigester interface {
		io.Writer
		BlobIdGetter
	}

	ReadDigester interface {
		io.Reader
		BlobIdGetter
	}

	ReadCloseDigester interface {
		io.WriterTo
		io.ReadCloser
		BlobIdGetter
	}

	WriteCloseDigester interface {
		io.ReaderFrom
		io.WriteCloser
		BlobIdGetter
	}
)
