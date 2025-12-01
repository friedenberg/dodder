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
