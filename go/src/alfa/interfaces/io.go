package interfaces

import (
	"io"
)

type FuncReader func(io.Reader) (int64, error)

type (
	FuncReaderFormat[T any]  func(io.Reader, *T) (int64, error)
	FuncWriterElement[T any] func(io.Writer, *T) (int64, error)

	// TODO-P3 switch to below
	FuncReaderFormatInterface[T any]  func(io.Reader, T) (int64, error)
	FuncReaderElementInterface[T any] func(io.Writer, T) (int64, error)
	FuncWriterElementInterface[T any] func(io.Writer, T) (int64, error)
)

type (
	ReadWrapper interface {
		WrapReader(r io.Reader) (io.ReadCloser, error)
	}

	WriteWrapper interface {
		WrapWriter(w io.Writer) (io.WriteCloser, error)
	}

	IOWrapper interface {
		ReadWrapper
		WriteWrapper
	}

	WriterAndStringWriter interface {
		io.Writer
		io.StringWriter
	}

	FuncWriter              func(io.Writer) (int64, error)
	FuncWriterFormat[T any] func(io.Writer, T) (int64, error)

	FuncStringWriterFormat[T any] func(WriterAndStringWriter, T) (int64, error)

	FuncMakePrinter[OUT any] func(WriterAndStringWriter) FuncIter[OUT]
)
