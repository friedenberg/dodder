package triple_hyphen_io2

import "io"

type MetadataWriterTo interface {
	io.WriterTo
	HasMetadataContent() bool
}

type readerState int

const (
	readerStateEmpty = readerState(iota)
	readerStateFirstBoundary
	readerStateSecondBoundary
)
