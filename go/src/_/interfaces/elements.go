package interfaces

type ValueLike interface {
	Stringer
	EqualsAny(any) bool
}

type Lessor[ELEMENT any] interface {
	Less(ELEMENT, ELEMENT) bool
}

// TODO-P2 rename
type Equaler[ELEMENT any] interface {
	Equals(ELEMENT, ELEMENT) bool
}

type ResetterPtr[
	ELEMENT any,
	ELEMENT_PTR Ptr[ELEMENT],
] interface {
	Reset(ELEMENT_PTR)
	ResetWith(ELEMENT_PTR, ELEMENT_PTR)
}

type Resetter[ELEMENT any] interface {
	Reset(ELEMENT)
	ResetWith(ELEMENT, ELEMENT)
}

type Equatable[ELEMENT any] interface {
	Equals(ELEMENT) bool
}

type Resetable interface {
	Reset()
}

type ResetableWithError interface {
	Reset() error
}

type ResetablePtr[ELEMENT any] interface {
	Ptr[ELEMENT]
	ResetWith(ELEMENT)
	Reset()
}
