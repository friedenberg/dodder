package interfaces

import "iter"

type (
	Seq[T any]          = iter.Seq[T]
	Seq2[T any, T1 any] = iter.Seq2[T, T1]
	SeqError[T any]     = Seq2[T, error]
)

func MakeSeqErrorWithError[T any](err error) SeqError[T] {
	return func(yield func(T, error) bool) {
		var t T
		yield(t, err)
	}
}
