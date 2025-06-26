package quiter

import (
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func ResetMap[K comparable, V any](m map[K]V) {
	for k := range m {
		delete(m, k)
	}
}

func ResetMutableSetWithPool[E any, EPtr interfaces.Ptr[E]](
	s interfaces.MutableSetPtrLike[E, EPtr],
	p interfaces.Pool[E, EPtr],
) {
	s.EachPtr(p.Put)
	s.Reset()
}
