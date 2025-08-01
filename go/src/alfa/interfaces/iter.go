package interfaces

import (
	"iter"
)

type (
	Seq[T any]          = iter.Seq[T]
	Seq2[T any, T1 any] = iter.Seq2[T, T1]
	SeqError[T any]     = iter.Seq2[T, error]
)
