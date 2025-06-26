package collections

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func WriterContainer[T interfaces.Element](
	s interfaces.SetLike[T],
	sigil error,
) interfaces.FuncIter[T] {
	return func(e T) (err error) {
		k := s.Key(e)

		if k == "" {
			err = ErrEmptyKey[T]{Element: e}
			return
		}

		_, ok := s.Get(k)

		if !ok {
			err = errors.Wrap(sigil)
		}

		return
	}
}

// func WriterFuncNegate[T any](wf schnittstellen.FuncIter[T])
// schnittstellen.FuncIter[T] {
// 	return func(e T) (err error) {
// 		err = wf(e)

// 		switch {
// 		case err == nil:
// 			err = MakeErrStopIteration()

// 		case IsStopIteration(err):
// 			err = nil
// 		}

// 		return
// 	}
// }

// func (s1 Set[T]) Subtract(s2 Set[T]) (out Set[T]) {
// 	s3 := makeSet[T](s1.Key)
// 	s3.open()
// 	defer s3.close()

// 	s1.Chain(
// 		WriterFuncNegate(s2.WriterContainer(MakeErrStopIteration())),
// 		s3.add,
// 	)

// 	out.Set = s3

// 	return
// }

// func (s1 Set[T]) Intersection(
// 	s2 schnittstellen.Set[T],
// ) (s3 schnittstellen.MutableSet[T]) {
// 	s3 = MakeMutableSet[T](s1.Key)
// 	s22 := Set[T]{
// 		Set: s2,
// 	}

// 	s1.Chain(
// 		s22.WriterContainer(MakeErrStopIteration()),
// 		s3.Add,
// 	)

// 	return
// }

// func (s1 Set[T]) Chain(fs ...schnittstellen.FuncIter[T]) error {
// 	return s1.Each(
// 		func(e T) (err error) {
// 			for _, f := range fs {
// 				if err = f(e); err != nil {
// 					if IsStopIteration(err) {
// 						err = nil
// 					} else {
// 						err = errors.Wrap(err)
// 					}

// 					return
// 				}
// 			}

// 			return
// 		},
// 	)
// }

// func (s Set[T]) Elements() (out []T) {
// 	out = make([]T, 0, s.Len())

// 	s.Each(
// 		func(e T) (err error) {
// 			out = append(out, e)
// 			return
// 		},
// 	)

// 	return
// }

// func (s Set[T]) Any() (e T) {
// 	s.Each(
// 		func(e1 T) (err error) {
// 			e = e1
// 			return MakeErrStopIteration()
// 		},
// 	)

// 	return
// }

// func (s Set[T]) All(f schnittstellen.FuncIter[T]) (ok bool) {
// 	err := s.Each(
// 		func(e T) (err error) {
// 			return f(e)
// 		},
// 	)

// 	return err == nil
// }

// func (a Set[T]) Equals(b schnittstellen.Set[T]) (ok bool) {
// 	if a.Len() != b.Len() {
// 		return
// 	}

// 	ok = a.All(Set[T]{Set: b}.WriterContainer(ErrNotFound{}))

// 	return
// }

// func (outer Set[T]) ContainsSet(inner Set[T]) (ok bool) {
// 	if outer.Len() < inner.Len() {
// 		return
// 	}

// 	ok = inner.All(outer.WriterContainer(ErrNotFound{}))

// 	return
// }
