package interfaces

import "io"

// TODO-P3 refactor into hash or checksum or content address and split korper
// out into context object
type Sha interface {
	// TODO-P3
	// GetHashBytes() []byte
	// ValueLike
	StringerWithHeadAndTail
	GetShaString() string
	GetShaBytes() []byte
	EqualsSha(Sha) bool // TODO-P3 rename to EqualsShaLike
	IsNull() bool
	ShaGetter
}

type ShaGetter interface {
	GetShaLike() Sha
}

// TODO reconsider this and force consumption of bufio? Formats expect
// WriterAndStringWriter, but this forces just Writer
type (
	// TODO rename to BlobReader
	ShaReadCloser interface {
		io.WriterTo
		io.ReadCloser
		GetShaLike() Sha
	}

	// TODO rename to BlobWriter
	ShaWriteCloser interface {
		io.ReaderFrom
		io.WriteCloser
		GetShaLike() Sha
	}

	ShaWriter interface {
		WriterAndStringWriter
		ShaGetter
	}
)

type (
	FuncShaReadCloser  func(Sha) (ShaReadCloser, error)
	FuncShaWriteCloser func(Sha) (ShaWriteCloser, error)
)
