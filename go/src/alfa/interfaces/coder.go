package interfaces

import (
	"bufio"
	"io"
)

type (
	DecoderFrom[B any, R any] interface {
		DecodeFrom(B, R) (int64, error)
	}

	EncoderTo[B any, R any] interface {
		EncodeTo(B, R) (int64, error)
	}

	Coder[B any, R any, W any] interface {
		DecoderFrom[B, R]
		EncoderTo[B, W]
	}

	DecoderFromBufferedReader[B any] = DecoderFrom[B, *bufio.Reader]
	EncoderToBufferedWriter[B any]   = EncoderTo[B, *bufio.Writer]
	CoderBufferedReadWriter[B any]   = Coder[B, *bufio.Reader, *bufio.Writer]

	DecoderFromReader[B any] = DecoderFrom[B, io.Reader]
	EncoderToWriter[B any]   = EncoderTo[B, io.Writer]
	CoderReadWriter[B any]   = Coder[B, io.Reader, io.Writer]

	StringEncoderTo[T any] interface {
		EncodeStringTo(T, WriterAndStringWriter) (int64, error)
	}

	StringCoder[T any] interface {
		DecoderFromReader[T]
		StringEncoderTo[T]
	}
)
