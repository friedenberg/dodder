package interfaces

import (
	"bytes"
	"fmt"
	"io"
)

type (
	Digest interface {
		DigestGetter
		GetBytes() []byte
		GetType() string
		IsNull() bool
	}

	DigestGetter interface {
		GetDigest() Digest
	}

	ReadCloserDigester interface {
		io.WriterTo
		io.ReadCloser
		DigestGetter
	}

	WriteCloserDigester interface {
		io.ReaderFrom
		io.WriteCloser
		DigestGetter
	}
)

func DigestEquals(a, b Digest) bool {
	return bytes.Equal(a.GetBytes(), b.GetBytes())
}

// Creates a human-readable string representation of a digest.
// TODO add type information
func FormatDigest(digest Digest) string {
	return fmt.Sprintf("%x", digest.GetBytes())
}
