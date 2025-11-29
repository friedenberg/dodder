package collections_value

import (
	"encoding/gob"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func MakeValueSetString[
	T interfaces.Stringer,
	TPtr interfaces.StringSetterPtr[T],
](
	keyer interfaces.StringKeyer[T],
	es ...string,
) (s Set[T], err error) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyer[T]{}
	}

	s.K = keyer

	for _, v := range es {
		var e T
		e1 := TPtr(&e)

		if err = e1.Set(v); err != nil {
			err = errors.Wrap(err)
			return s, err
		}

		s.E[s.K.GetKey(e)] = e
	}

	return s, err
}

func MakeValueSetValue[T interfaces.Stringer](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s Set[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyer[T]{}
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeValueSet[ELEMENT interfaces.Stringer](
	keyer interfaces.StringKeyer[ELEMENT],
	seq interfaces.Seq[ELEMENT],
) (set Set[ELEMENT]) {
	gob.Register(set)
	set.E = make(map[string]ELEMENT, 0)

	if keyer == nil {
		keyer = quiter.StringerKeyer[ELEMENT]{}
	}

	set.K = keyer

	for element := range seq {
		set.E[set.K.GetKey(element)] = element
	}

	return set
}

func MakeValueSetFromSlice[ELEMENT interfaces.Stringer](
	keyer interfaces.StringKeyer[ELEMENT],
	elements ...ELEMENT,
) (set Set[ELEMENT]) {
	gob.Register(set)
	set.E = make(map[string]ELEMENT, len(elements))

	if keyer == nil {
		keyer = quiter.StringerKeyer[ELEMENT]{}
	}

	set.K = keyer

	for index := range elements {
		element := elements[index]
		set.E[set.K.GetKey(element)] = element
	}

	return set
}

func MakeSetValue[T interfaces.Stringer](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s Set[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeSet[T any](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s Set[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeMutableValueSetValue[T interfaces.Stringer](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s MutableSet[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyer[T]{}
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeMutableValueSet[T interfaces.Stringer](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s MutableSet[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyer[T]{}
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeMutableSetValue[T any](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s MutableSet[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}

func MakeMutableSet[T any](
	keyer interfaces.StringKeyer[T],
	es ...T,
) (s MutableSet[T]) {
	gob.Register(s)
	s.E = make(map[string]T, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKey(e)] = e
	}

	return s
}
