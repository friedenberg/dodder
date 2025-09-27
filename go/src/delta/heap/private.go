package heap

import (
	"container/heap"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

type Element interface{}

type ElementPtr[T Element] interface {
	interfaces.Ptr[T]
}

type heapPrivate[T Element, TPtr ElementPtr[T]] struct {
	Lessor     interfaces.Lessor3[TPtr]
	Resetter   interfaces.Resetter2[T, TPtr]
	Elements   []TPtr
	lastPopped TPtr
	p          interfaces.Pool[T, TPtr]
	equaler    interfaces.Equaler[TPtr]
}

func (h *heapPrivate[T, TPtr]) GetPool() interfaces.Pool[T, TPtr] {
	if h.p == nil {
		h.p = pool.MakeFakePool[T, TPtr]()
	}

	return h.p
}

func (h heapPrivate[T, TPtr]) Len() int {
	return len(h.Elements)
}

func (h heapPrivate[T, TPtr]) Less(i, j int) (ok bool) {
	ok = h.Lessor.Less(h.Elements[i], h.Elements[j])
	return ok
}

func (h heapPrivate[T, TPtr]) Swap(i, j int) {
	h.Elements[i], h.Elements[j] = h.Elements[j], h.Elements[i]
}

func (h *heapPrivate[T, TPtr]) Push(x any) {
	h.Elements = append(h.Elements, x.(TPtr))
}

func (h *heapPrivate[T, TPtr]) discardDupes() {
	panic("don't use this yet")

	for h.lastPopped != nil &&
		h.Len() > 0 &&
		h.equaler.Equals(h.lastPopped, h.Elements[0]) {
		heap.Pop(h)
		// d := heap.Pop(h)
		// log.Debug().Printf("discarded: %s", d)
	}
}

func (h *heapPrivate[T, TPtr]) Pop() any {
	old := h.Elements
	n := len(old)
	x := old[n-1]
	h.Elements = old[0 : n-1]

	return x
}

func (h *heapPrivate[T, TPtr]) saveLastPopped(e TPtr) {
	if h.lastPopped == nil {
		h.lastPopped = h.GetPool().Get()
	}

	h.Resetter.ResetWith(h.lastPopped, e)
}

func (a heapPrivate[T, TPtr]) Copy() (b heapPrivate[T, TPtr]) {
	b = heapPrivate[T, TPtr]{
		Lessor:   a.Lessor,
		equaler:  a.equaler,
		p:        a.p,
		Resetter: a.Resetter,
		Elements: make([]TPtr, a.Len()),
	}

	copy(b.Elements, a.Elements)

	return b
}

func (a heapPrivate[T, TPtr]) Sorted() (b heapPrivate[T, TPtr]) {
	b = a.Copy()
	sort.Sort(b)
	return b
}
