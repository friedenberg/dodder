package interfaces

import (
	"hash"
	"io"
)

type (
	Digest interface {
		Digester
		GetBytes() []byte
		GetType() string
		IsNull() bool
	}

	MutableDigest interface {
		Digester
		SetBytes([]byte) error
		Reset()
	}

	Digester interface {
		GetDigest() Digest
	}

	EnvDigest interface {
		GetType() string

		// TODO rename to "MakeHash"
		GetHash() hash.Hash
		PutHash(hash.Hash)

		// TODO rename to "MakeDigest"
		GetDigest() Digest
		PutDigest(Digest)

		MakeWriteDigesterWithRepool() (WriteDigester, func())
		MakeWriteDigester() WriteDigester
		MakeDigestFromHash(hash.Hash) (Digest, error)

		// TODO pool
		// GetDigest() Digest
		// PutDigest(Digest)
	}

	WriteDigester interface {
		io.Writer
		Digester
	}

	ReadDigester interface {
		io.Reader
		Digester
	}

	ReadCloseDigester interface {
		io.WriterTo
		io.ReadCloser
		Digester
	}

	WriteCloseDigester interface {
		io.ReaderFrom
		io.WriteCloser
		Digester
	}
)
