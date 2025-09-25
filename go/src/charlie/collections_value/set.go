package collections_value

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Set[
	T any,
] struct {
	K interfaces.StringKeyer[T]
	E map[string]T
}

func (s Set[T]) AllKeys() interfaces.Seq[string] {
	return func(yield func(string) bool) {
		for k := range s.E {
			if !yield(k) {
				break
			}
		}
	}
}

func (s Set[T]) All() interfaces.Seq[T] {
	return func(yield func(T) bool) {
		for _, e := range s.E {
			if !yield(e) {
				break
			}
		}
	}
}

func (s Set[T]) Len() int {
	if s.E == nil {
		return 0
	}

	return len(s.E)
}

func (s Set[T]) Key(e T) string {
	return s.K.GetKey(e)
}

func (s Set[T]) Get(k string) (e T, ok bool) {
	e, ok = s.E[k]

	return e, ok
}

func (s Set[T]) Any() (e T) {
	for _, e1 := range s.E {
		return e1
	}

	return e
}

func (s Set[T]) ContainsKey(k string) (ok bool) {
	if k == "" {
		return ok
	}

	_, ok = s.E[k]

	return ok
}

func (s Set[T]) Contains(e T) (ok bool) {
	return s.ContainsKey(s.Key(e))
}

// TODO remove in favor of iterators
func (s Set[T]) EachKey(wf interfaces.FuncIterKey) (err error) {
	for v := range s.E {
		if err = wf(v); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}
	}

	return err
}

func (s Set[T]) Add(v T) (err error) {
	s.E[s.Key(v)] = v
	return err
}

func (a Set[T]) CloneSetLike() interfaces.SetLike[T] {
	return a
}

func (a Set[T]) CloneMutableSetLike() interfaces.MutableSetLike[T] {
	c := MakeMutableSet[T](a.K)
	for e := range a.All() {
		c.Add(e)
	}
	return c
}
