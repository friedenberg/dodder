package interfaces

import (
	"encoding"
	"hash"
	"io"
)

type (
	MarklId interface {
		// TODO consider removing Stringer and Setter
		Stringer
		encoding.BinaryMarshaler
		encoding.TextMarshaler
		// io.WriterTo
		GetBytes() []byte
		// TODO rethink size as it works completely different between sha and
		// merkle
		GetSize() int
		GetType() string
		IsNull() bool
	}

	MutableMarklId interface {
		MarklId
		Setter
		encoding.BinaryUnmarshaler
		encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMerkleId(tipe string, bites []byte) error
		Reset()
		ResetWithMarklId(MarklId)
	}

	MarklIdGetter interface {
		GetMarklId() MarklId
	}

	Hash interface {
		hash.Hash
		GetType() string
		GetMarklId() (MutableMarklId, FuncRepool)
	}

	HashType interface {
		GetMarklIdForString(input string) (MarklId, FuncRepool)
		// TODO rename
		FromStringFormat(format string, args ...any) (MarklId, FuncRepool)
	}

	WriteMarklIdGetter interface {
		io.Writer
		MarklIdGetter
	}

	ReadMarklIdGetter interface {
		io.Reader
		MarklIdGetter
	}

	ReadCloseMarklIdGetter interface {
		io.WriterTo
		io.ReadCloser
		MarklIdGetter
	}

	WriteCloseMarklIdGetter interface {
		io.ReaderFrom
		io.WriteCloser
		MarklIdGetter
	}
)
