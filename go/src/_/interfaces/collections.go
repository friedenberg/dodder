package interfaces

type (
	FuncIter[ELEMENT any] func(ELEMENT) error

	FuncIterIO[ELEMENT any]            func(ELEMENT) (int64, error)
	FuncTransform[ELEMENT any, T1 any] func(ELEMENT) (T1, error)
	FuncIterKey                        func(string) error
	FuncIterWithKey[ELEMENT any]       func(string, ELEMENT) error
)

type (
	Lenner interface {
		Len() int
	}

	ContainsKeyer interface {
		ContainsKey(string) bool
	}

	Container[ELEMENT any] interface {
		Contains(ELEMENT) bool
	}

	Aller[ELEMENT any] interface {
		All() Seq[ELEMENT]
	}

	Iterable[ELEMENT any] interface {
		Any() ELEMENT // TODO make derived
		All() Seq[ELEMENT]
	}

	Adder[E any] interface {
		Add(E) error
	}

	AdderPtr[E any, EPtr Ptr[E]] interface {
		AddPtr(EPtr) error
	}

	SetBase[ELEMENT any] interface {
		Lenner
		// Container[ELEMENT]
		ContainsKeyer
		Aller[ELEMENT]
	}
)

type Delta[T any] interface {
	GetAdded() SetLike[T]
	GetRemoved() SetLike[T]
}

type (
	Collection[T any] interface {
		Lenner
		Iterable[T]
	}

	CollectionPtr[T any, TPtr Ptr[T]] interface {
		Lenner
	}
)

type (
	SetLike[T any] interface {
		Collection[T]
		ContainsKeyer

		Key(T) string
		Get(string) (T, bool)
		Contains(T) bool

		CloneSetLike() SetLike[T]
		CloneMutableSetLike() MutableSetLike[T]
	}

	SetPtrLike[T any, TPtr Ptr[T]] interface {
		SetLike[T]
		CollectionPtr[T, TPtr]

		GetPtr(string) (TPtr, bool)
		KeyPtr(TPtr) string

		CloneSetPtrLike() SetPtrLike[T, TPtr]
		CloneMutableSetPtrLike() MutableSetPtrLike[T, TPtr]
	}

	MutableSetPtrLike[T any, TPtr Ptr[T]] interface {
		SetPtrLike[T, TPtr]
		MutableSetLike[T]
		AddPtr(TPtr) error
		DelPtr(TPtr) error
	}

	MutableSetLike[T any] interface {
		SetLike[T]
		Adder[T]
		Del(T) error
		DelKey(string) error
		Resetable
	}
)

type (
	TridexLike interface {
		Collection[string]
		ContainsAbbreviation(string) bool
		ContainsExpansion(string) bool
		Abbreviate(string) string
		Expand(string) string
	}

	MutableTridexLike interface {
		TridexLike
		Add(string)
		Remove(string)
	}

	Tridex interface {
		TridexLike
	}

	MutableTridex interface {
		Tridex
		Add(string)
		Remove(string)
	}
)
