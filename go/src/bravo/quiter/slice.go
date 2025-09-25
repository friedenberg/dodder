package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
)

// TODO move to own package
type Slice[E any] []E

func (s Slice[E]) Len() int {
	return len(s)
}

func (s Slice[E]) Any() (e E) {
	if s.Len() > 0 {
		e = s[0]
	}

	return e
}

func (s Slice[E]) Each(f interfaces.FuncIter[E]) error {
	return comments.Implement()
}

func (s Slice[E]) All() interfaces.Seq[E] {
	return func(yield func(E) bool) {
		for _, e := range s {
			if !yield(e) {
				break
			}
		}
	}
}

func (s *Slice[E]) Append(element ...E) {
	*s = append(*s, element...)
}
