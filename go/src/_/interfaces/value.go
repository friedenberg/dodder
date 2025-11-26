package interfaces

type (
	Value[SELF any] interface {
		Stringer
		Equatable[SELF]
		IsEmpty() bool
	}

	ValuePtr[SELF Value[SELF]] interface {
		Resetable
		ResetablePtr[SELF]
		StringerSetterPtr[SELF]
	}
)

// type Value[T any] interface {
// 	ValueLike
// 	Equatable[T]
// }

// type ValuePtr[T any] interface {
// 	ValueLike
// 	// Value[T]
// 	Ptr[T]
// }
