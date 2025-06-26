package interfaces

type Element interface {
	EqualsAny(any) bool
}

type ElementPtr[T any] interface {
	Ptr[T]
	Element
}

type ValueLike interface {
	Stringer
	Element
}

type Value[T any] interface {
	ValueLike
	Equatable[T]
}

type ValuePtr[T any] interface {
	ValueLike
	// Value[T]
	Ptr[T]
}

// TODO-P2 remove
type Lessor[T any] interface {
	Less(T) bool
}

type Lessor3[T any] interface {
	Less(T, T) bool
}

// TODO-P2 rename
type Equaler[T any] interface {
	Equals(T, T) bool
}

type Resetter2[T any, TPtr Ptr[T]] interface {
	Reset(TPtr)
	ResetWith(TPtr, TPtr)
}

type Resetter3[T any] interface {
	Reset(T)
	ResetWith(T, T)
}
