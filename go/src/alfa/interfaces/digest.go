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
		GetHash() hash.Hash
		PutHash(hash.Hash)

		GetDigest() Digest
		PutDigest(Digest)

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
