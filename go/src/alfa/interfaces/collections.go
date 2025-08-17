package interfaces

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
	Collection[T]
	ContainsKeyer

	Key(T) string
	Get(string) (T, bool)
	Contains(T) bool
	AllKeys() Seq[string]

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

type Adder[E any] interface {
	Add(E) error
}

type AdderPtr[E any, EPtr Ptr[E]] interface {
	AddPtr(EPtr) error
}
