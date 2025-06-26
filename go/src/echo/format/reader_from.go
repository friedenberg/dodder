package format

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type readerFrom[T any] struct {
	rf interfaces.FuncReaderFormat[T]
	e  *T
}

func (rf readerFrom[T]) ReadFrom(r io.Reader) (n int64, err error) {
	return rf.rf(r, rf.e)
}

func MakeReaderFrom[T any](
	rf interfaces.FuncReaderFormat[T],
	e *T,
) io.ReaderFrom {
	return readerFrom[T]{
		rf: rf,
		e:  e,
	}
}

type readerFromInterface[T any] struct {
	rf interfaces.FuncReaderFormatInterface[T]
	e  T
}

func (rf readerFromInterface[T]) ReadFrom(r io.Reader) (n int64, err error) {
	return rf.rf(r, rf.e)
}

func MakeReaderFromInterface[T any](
	rf interfaces.FuncReaderFormatInterface[T],
	e T,
) io.ReaderFrom {
	return readerFromInterface[T]{
		rf: rf,
		e:  e,
	}
}
