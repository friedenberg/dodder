package interfaces

import (
	"encoding"
	"hash"
	"io"
)

type (
	MarklType interface {
		GetMarklTypeId() string
	}

	HashType interface {
		MarklType
		GetMarklIdForString(input string) (MarklId, FuncRepool)
		// TODO rename
		FromStringFormat(format string, args ...any) (MarklId, FuncRepool)
	}

	MarklTypeGetter interface {
		GetMarklType() MarklType
	}

	Hash interface {
		hash.Hash
		MarklTypeGetter
		GetMarklId() (MutableMarklId, FuncRepool)
	}

	MarklId interface {
		// TODO consider removing Stringer and Setter
		Stringer
		StringWithFormat() string
		encoding.BinaryMarshaler
		// encoding.TextMarshaler
		// io.WriterTo
		GetBytes() []byte
		// TODO rethink size as it works completely different between sha and
		// merkle
		GetSize() int
		MarklTypeGetter
		IsNull() bool
		GetFormat() string
	}

	MutableMarklId interface {
		MarklId
		Setter
		encoding.BinaryUnmarshaler
		// encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMerkleId(typeId string, bites []byte) error
		Reset()
		ResetWithMarklId(MarklId)
		SetFormat(string) error
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
