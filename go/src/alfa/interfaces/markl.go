package interfaces

import (
	"encoding"
	"hash"
	"io"
)

type (
	HashType interface {
		GetType() string
		GetMarklIdForString(input string) (MarklId, FuncRepool)
		// TODO rename
		FromStringFormat(format string, args ...any) (MarklId, FuncRepool)
	}

	Hash interface {
		hash.Hash
		GetType() HashType
		GetMarklId() (MutableMarklId, FuncRepool)
	}

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
		GetType() HashType
		IsNull() bool
	}

	MutableMarklId interface {
		MarklId
		Setter
		encoding.BinaryUnmarshaler
		encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMerkleId(typeId string, bites []byte) error
		Reset()
		ResetWithMarklId(MarklId)
	}

	MarklIdGetter interface {
		GetMarklId() MarklId
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
