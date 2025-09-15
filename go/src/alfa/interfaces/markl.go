package interfaces

import (
	"encoding"
	"hash"
)

type (
	MarklFormat interface {
		GetSize() int
		GetMarklFormatId() string
	}

	FormatHash interface {
		MarklFormat

		GetHash() Hash
		PutHash(Hash)

		GetMarklIdForString(input string) (MarklId, FuncRepool)
		GetMarklIdFromStringFormat(
			format string,
			args ...any,
		) (MarklId, FuncRepool)
	}

	MarklFormatGetter interface {
		GetMarklFormat() MarklFormat
	}

	Hash interface {
		hash.Hash
		MarklFormatGetter
		GetMarklId() (MutableMarklId, FuncRepool)
	}

	MarklId interface {
		// TODO consider removing Stringer and Setter

		// TODO add WriteString and WriteStringWithFormat
		Stringer
		StringWithFormat() string

		encoding.BinaryMarshaler
		// encoding.TextMarshaler
		// io.WriterTo
		GetBytes() []byte
		// TODO rethink size as it works completely different between sha and
		// merkle
		GetSize() int
		MarklFormatGetter
		IsNull() bool

		GetPurpose() string

		// Optional methods
		GetIOWrapper() (IOWrapper, error)
		Verify(mes, sig MarklId) error
	}

	MutableMarklId interface {
		MarklId
		Setter
		encoding.BinaryUnmarshaler
		// encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMarklId(formatId string, bites []byte) error
		Reset()
		ResetWithMarklId(MarklId)
		SetPurpose(string) error
	}

	MarklIdGetter interface {
		GetMarklId() MarklId
	}
)
