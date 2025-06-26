package interfaces

import (
	"iter"
)

type Delta[T any] interface {
	GetAdded() SetLike[T]
	GetRemoved() SetLike[T]
}

type CollectionOld[T any] interface {
	Lenner
	Iterable[T]
	Each(FuncIter[T]) error // TODO remove in favor of iter.Seq
}

type Collection[T any] interface {
	Lenner
	Iterable[T]
}

type SetLike[T any] interface {
	CollectionOld[T]
	ContainsKeyer

	Key(T) string
	Get(string) (T, bool)
	Contains(T) bool
	AllKeys() iter.Seq[string]

	CloneSetLike() SetLike[T]
	CloneMutableSetLike() MutableSetLike[T]
}

type MutableSetLike[T any] interface {
	SetLike[T]
	Adder[T]
	Del(T) error
	DelKey(string) error
	Resetter
}

type TridexLike interface {
	Lenner
	EachString(FuncIter[string]) error
	ContainsAbbreviation(string) bool
	ContainsExpansion(string) bool
	Abbreviate(string) string
	Expand(string) string
}

type MutableTridexLike interface {
	TridexLike
	Add(string)
	Remove(string)
}

type Tridex interface {
	TridexLike
}

type MutableTridex interface {
	Tridex
	Add(string)
	Remove(string)
}

type Poolable[T any] interface{}

type PoolablePtr[T any] interface {
	Ptr[T]
	// Resetable[T]
}

type PoolValue[T any] interface {
	Get() T
	Put(i T) (err error)
}

type Pool[T Poolable[T], TPtr PoolablePtr[T]] interface {
	PoolValue[TPtr]
	PutMany(...TPtr) error
}

type PoolValue2[T any] interface {
	Get() (T, error)
	Put(i T) (err error)
}

type Pool2[T Poolable[T], TPtr PoolablePtr[T]] interface {
	PoolValue2[TPtr]
	PutMany(...TPtr) error
}

type Adder[E any] interface {
	Add(E) error
}

type AdderPtr[E any, EPtr Ptr[E]] interface {
	AddPtr(EPtr) error
}
