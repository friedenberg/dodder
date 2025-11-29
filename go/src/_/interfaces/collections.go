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

	// TODO make derived
	Container[ELEMENT any] interface {
		Contains(ELEMENT) bool
	}

	Keyer[ELEMENT any] interface {
		Key(ELEMENT) string
	}

	Aller[ELEMENT any] interface {
		All() Seq[ELEMENT]
	}

	Iterable[ELEMENT any] interface {
		All() Seq[ELEMENT]
	}

	Adder[ELEMENT any] interface {
		Add(ELEMENT) error
	}

	AdderPtr[ELEMENT any, ELEMENT_PTR Ptr[ELEMENT]] interface {
		AddPtr(ELEMENT_PTR) error
	}

	SetBase[ELEMENT any] interface {
		Lenner
		Keyer[ELEMENT]
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
	SetLike[ELEMENT any] interface {
		Collection[ELEMENT]
		ContainsKeyer

		Key(ELEMENT) string
		Get(string) (ELEMENT, bool)
		Contains(ELEMENT) bool

		CloneMutableSetLike() MutableSetLike[ELEMENT]
	}

	SetPtrLike[ELEMENT any, ELEMENT_PTR Ptr[ELEMENT]] interface {
		SetLike[ELEMENT]
		CollectionPtr[ELEMENT, ELEMENT_PTR]

		GetPtr(string) (ELEMENT_PTR, bool)
		KeyPtr(ELEMENT_PTR) string

		CloneSetPtrLike() SetPtrLike[ELEMENT, ELEMENT_PTR]
		CloneMutableSetPtrLike() MutableSetPtrLike[ELEMENT, ELEMENT_PTR]
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
