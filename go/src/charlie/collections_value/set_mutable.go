package collections_value

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type MutableSet[
	T any,
] struct {
	K interfaces.StringKeyer[T]
	E map[string]T
}

func (s MutableSet[T]) Len() int {
	if s.E == nil {
		return 0
	}

	return len(s.E)
}

func (s MutableSet[T]) All() interfaces.Seq[T] {
	return func(yield func(T) bool) {
		for _, e := range s.E {
			if !yield(e) {
				break
			}
		}
	}
}

func (s MutableSet[T]) AllKeys() interfaces.Seq[string] {
	return func(yield func(string) bool) {
		for k := range s.E {
			if !yield(k) {
				break
			}
		}
	}
}

func (s MutableSet[T]) Key(e T) string {
	return s.K.GetKey(e)
}

func (s MutableSet[T]) Get(k string) (e T, ok bool) {
	e, ok = s.E[k]

	return
}

func (s MutableSet[T]) ContainsKey(k string) (ok bool) {
	if k == "" {
		return
	}

	_, ok = s.E[k]

	return
}

func (s MutableSet[T]) Contains(e T) (ok bool) {
	return s.ContainsKey(s.Key(e))
}

func (s MutableSet[T]) Any() (v T) {
	for _, v1 := range s.E {
		v = v1
		break
	}

	return
}

func (s MutableSet[T]) Del(v T) (err error) {
	return s.DelKey(s.Key(v))
}

func (s MutableSet[T]) DelKey(k string) (err error) {
	delete(s.E, k)
	return
}

func (s MutableSet[T]) Add(v T) (err error) {
	s.E[s.Key(v)] = v
	return
}

func (s MutableSet[T]) EachKey(
	wf interfaces.FuncIterKey,
) (err error) {
	for v := range s.E {
		if err = wf(v); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return
		}
	}

	return
}

func (a MutableSet[T]) Reset() {
	for k := range a.E {
		delete(a.E, k)
	}
}

func (a MutableSet[T]) CloneSetLike() interfaces.SetLike[T] {
	b := MakeSet[T](a.K)

	for k, v := range a.E {
		b.E[k] = v
	}

	return b
}

func (a MutableSet[T]) CloneMutableSetLike() interfaces.MutableSetLike[T] {
	c := MakeMutableSet[T](a.K)
	for e := range a.All() {
		c.Add(e)
	}
	return c
}
