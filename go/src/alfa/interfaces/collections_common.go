package interfaces

import "code.linenisgreat.com/zit/go/zit/src/alfa/interfaces"

type (
	// Yield[T any]           = func(T) bool
	// Yield2[T1 any, T2 any] = func(T1, T2) bool
	// YieldError[T any]      = Yield2[T, error]
	// Seq[T any]             = iter.Seq[T]
	// Seq2[T1 any, T2 any]   = iter.Seq2[T1, T2]
	// SeqError[T any]        = Seq2[T, error]

	FuncIter[T any] func(T) error

	FuncIterIO[T any]            func(T) (int64, error)
	FuncTransform[T any, T1 any] func(T) (T1, error)
	FuncIterKey                  func(string) error
	FuncIterWithKey[T any]       func(string, T) error
)

type Lenner interface {
	Len() int
}

type ContainsKeyer interface {
	ContainsKey(string) bool
}

type Iterable[T any] interface {
	Any() T
	All() interfaces.Seq[T]
}

type IterablePtr[T any, TPtr Ptr[T]] interface {
	AllPtr() interfaces.Seq[TPtr]
}
