package interfaces

import (
	"bytes"
	"fmt"
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

func DigesterEquals(a, b Digester) bool {
	return DigestEquals(a.GetDigest(), b.GetDigest())
}

func DigestEquals(a, b Digest) bool {
	return bytes.Equal(a.GetBytes(), b.GetBytes())
}

func FormatDigester(digester Digester) string {
	return FormatDigest(digester.GetDigest())
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func FormatDigest(digest Digest) string {
	return fmt.Sprintf("%x", digest.GetBytes())
}
