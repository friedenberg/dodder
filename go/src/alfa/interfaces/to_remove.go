package interfaces

type Equatable[T any] interface {
	Equals(T) bool
}

type Resetter interface {
	Reset()
}

type ResetterWithError interface {
	Reset() error
}

type Resetable[T any] interface {
	Ptr[T]
	ResetWith(T)
	Reset()
}
