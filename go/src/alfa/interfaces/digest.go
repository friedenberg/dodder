package interfaces

import "io"

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

	WriterStringWriterDigester interface {
		WriterAndStringWriter
		DigestGetter
	}
)
