package interfaces

import (
	"encoding"
	"hash"
	"io"
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
		// TODO add `WriteToMarklId` method for reuse
		GetMarklId() (MarklIdMutable, FuncRepool)
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
		IsEmpty() bool

		GetPurpose() string

		// Optional methods
		GetIOWrapper() (IOWrapper, error)
		Verify(mes, sig MarklId) error
		Sign(
			mes MarklId,
			sigDst MarklIdMutable,
			sigPurpose string,
		) (err error)
	}

	MarklIdMutable interface {
		MarklId
		Setter
		encoding.BinaryUnmarshaler
		// encoding.TextUnmarshaler
		// io.ReaderFrom
		SetMarklId(formatId string, bites []byte) error
		Reset()
		ResetWithMarklId(MarklId)
		SetPurpose(string) error

		// Optional methods
		GeneratePrivateKey(
			readerRand io.Reader,
			formatId string,
			purpose string,
		) (err error)
	}

	MarklIdGetter interface {
		GetMarklId() MarklId
	}

	DigestWriteMap map[string]MarklIdMutable

	Lock[
		KEY Value[KEY],
		KEY_PTR ValuePtr[KEY],
	] interface {
		GetKey() KEY
		GetValue() MarklId
		IsEmpty() bool
	}

	LockMutable[
		KEY Value[KEY],
		KEY_PTR ValuePtr[KEY],
	] interface {
		Lock[KEY, KEY_PTR]
		GetKeyMutable() KEY_PTR
		GetValueMutable() MarklIdMutable
	}
)
