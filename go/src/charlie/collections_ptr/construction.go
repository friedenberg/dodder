package collections_ptr

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func MakeValueSetString[
	T interfaces.Stringer,
	TPtr interfaces.StringSetterPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...string,
) (s Set[T, TPtr], err error) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}
	}

	s.K = keyer

	for _, v := range es {
		var e T
		e1 := TPtr(&e)

		if err = e1.Set(v); err != nil {
			err = errors.Wrap(err)
			return
		}

		s.E[s.K.GetKeyPtr(e1)] = e1
	}

	return
}

func MakeValueSetValue[
	T interfaces.Stringer,
	TPtr interfaces.StringerPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...T,
) (s Set[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}
	}

	s.K = keyer

	for i := range es {
		e := TPtr(&es[i])
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeValueSet[
	T interfaces.Stringer,
	TPtr interfaces.StringerPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...TPtr,
) (s Set[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeSetValue[T any, TPtr interfaces.Ptr[T]](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...T,
) (s Set[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := TPtr(&es[i])
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeSet[T any, TPtr interfaces.Ptr[T]](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...TPtr,
) (s Set[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeMutableValueSetValue[
	T interfaces.Stringer,
	TPtr interfaces.StringerPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...T,
) (s MutableSet[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}
	}

	s.K = keyer

	for i := range es {
		e := TPtr(&es[i])
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeMutableValueSet[
	T interfaces.Stringer,
	TPtr interfaces.StringerPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...TPtr,
) (s MutableSet[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		keyer = quiter.StringerKeyerPtr[T, TPtr]{}
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeMutableSetValue[
	T interfaces.Stringer,
	TPtr interfaces.StringerPtr[T],
](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...T,
) (s MutableSet[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := TPtr(&es[i])
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}

func MakeMutableSet[T any, TPtr interfaces.Ptr[T]](
	keyer interfaces.StringKeyerPtr[T, TPtr],
	es ...TPtr,
) (s MutableSet[T, TPtr]) {
	s.E = make(map[string]TPtr, len(es))

	if keyer == nil {
		panic("keyer was nil")
	}

	s.K = keyer

	for i := range es {
		e := es[i]
		s.E[s.K.GetKeyPtr(e)] = e
	}

	return
}
