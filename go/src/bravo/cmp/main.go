package cmp

import "code.linenisgreat.com/dodder/go/src/alfa/interfaces"

const (
	Less    = result(-1)
	Equal   = result(0)
	Greater = result(1)
)

type Func[ELEMENT any] func(ELEMENT, ELEMENT) Result

type Result interface {
	cmp()

	GetCompareInt() int
	Less() bool
	Equal() bool
	Greater() bool
}

type result int

var _ Result = result(0)

func (result) cmp() {}

func (result result) GetCompareInt() int {
	return int(result)
}

func (result result) Less() bool {
	return result == Less
}

func (result result) Equal() bool {
	return result == Equal
}

func (result result) Greater() bool {
	return result == Greater
}

func MakeFuncFromEqualerAndLessor3EqualFirst[ELEMENT any](
	equaler interfaces.Equaler[ELEMENT],
	lessor interfaces.Lessor3[ELEMENT],
) func(ELEMENT, ELEMENT) Result {
	return func(left, right ELEMENT) Result {
		if equaler.Equals(left, right) {
			return Equal
		} else if lessor.Less(left, right) {
			return Less
		} else {
			return Greater
		}
	}
}

func MakeFuncFromEqualerAndLessor3LessFirst[ELEMENT any](
	equaler interfaces.Equaler[ELEMENT],
	lessor interfaces.Lessor3[ELEMENT],
) func(ELEMENT, ELEMENT) Result {
	return func(left, right ELEMENT) Result {
		if lessor.Less(left, right) {
			return Less
		} else if equaler.Equals(left, right) {
			return Equal
		} else {
			return Greater
		}
	}
}

type Lesser[ELEMENT any] Func[ELEMENT]

func (lessor Lesser[T]) Less(left, right T) bool {
	return lessor(left, right).Less()
}

type Equaler[ELEMENT any] Func[ELEMENT]

func (equaler Equaler[T]) Equals(left, right T) bool {
	return equaler(left, right).Equal()
}
