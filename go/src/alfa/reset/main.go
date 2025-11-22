package reset

type (
	FuncReset[ELEMENT any]     func(ELEMENT)
	FuncResetWith[ELEMENT any] func(dst ELEMENT, src ELEMENT)
)

type resetter[ELEMENT any] struct {
	funcReset     FuncReset[ELEMENT]
	funcResetWith FuncResetWith[ELEMENT]
}

func MakeResetter[ELEMENT any](
	funcReset FuncReset[ELEMENT],
	funcResetWith FuncResetWith[ELEMENT],
) resetter[ELEMENT] {
	return resetter[ELEMENT]{
		funcReset:     funcReset,
		funcResetWith: funcResetWith,
	}
}

func (resetter resetter[ELEMENT]) Reset(element ELEMENT) {
	resetter.funcReset(element)
}

func (resetter resetter[ELEMENT]) ResetWith(dst, src ELEMENT) {
	resetter.funcResetWith(dst, src)
}
