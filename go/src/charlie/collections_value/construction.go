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

func MakeValueSet[T interfaces.Stringer](
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
