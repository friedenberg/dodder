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

	Keyer[ELEMENT any] interface {
		Key(ELEMENT) string
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

	Collection[ELEMENT any] interface {
		Lenner
		Iterable[ELEMENT]
	}

	SetGetter[ELEMENT any] interface {
		Get(string) (ELEMENT, bool)
	}

	Set[ELEMENT any] interface {
		ContainsKeyer
		Iterable[ELEMENT]
		Keyer[ELEMENT]
		Lenner
		SetGetter[ELEMENT]
	}
)

type Delta[ELEMENT any] interface {
	GetAdded() Set[ELEMENT]
	GetRemoved() Set[ELEMENT]
}

type (
	SetPtrLike[ELEMENT any, ELEMENT_PTR Ptr[ELEMENT]] interface {
		Set[ELEMENT]

		CloneSetPtrLike() SetPtrLike[ELEMENT, ELEMENT_PTR]
		CloneMutableSetPtrLike() MutableSetPtrLike[ELEMENT, ELEMENT_PTR]
	}

	MutableSetPtrLike[ELEMENT any, ELEMENT_PTR Ptr[ELEMENT]] interface {
		SetPtrLike[ELEMENT, ELEMENT_PTR]
		MutableSetLike[ELEMENT]
		AddPtr(ELEMENT_PTR) error
	}

	MutableSetLike[ELEMENT any] interface {
		Set[ELEMENT]
		Adder[ELEMENT]
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
