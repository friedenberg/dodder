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

	Digester interface {
		GetDigest() Digest
	}

	EnvDigest interface {
		MakeWriteDigester() WriteDigester
		MakeReadDigester() ReadDigester
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
