package interfaces

import (
	"io"
)

// TODO reconsider this and force consumption of bufio? Formats expect
// WriterAndStringWriter, but this forces just Writer
type (
	// TODO rename to BlobReader
	ShaReadCloser interface {
		io.WriterTo
		io.ReadCloser
		DigestGetter
	}

	// TODO rename to BlobWriter
	ShaWriteCloser interface {
		io.ReaderFrom
		io.WriteCloser
		DigestGetter
	}
)
