package heap

import (
	"container/heap"
	"iter"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

type Heap[T Element, TPtr ElementPtr[T]] struct {
	l sync.Mutex
	h heapPrivate[T, TPtr]
	s int
}

func (h *Heap[T, TPtr]) GetCollection() interfaces.Collection[TPtr] {
	return h
}

func (h *Heap[T, TPtr]) Any() TPtr {
	e, _ := h.Peek()
	return e
}

func (h *Heap[T, TPtr]) All() iter.Seq[TPtr] {
	return func(yield func(TPtr) bool) {
		h.l.Lock()
		defer h.l.Unlock()
		defer h.restore()

		for {
			e, ok := h.popAndSave()

			if !ok {
				return
			}

			if !yield(e) {
				break
			}
		}
	}
}

func (h *Heap[T, TPtr]) SetPool(v interfaces.Pool[T, TPtr]) {
	h.h.p = v
}

func (h *Heap[T, TPtr]) GetEqualer() interfaces.Equaler[TPtr] {
	return h.h.equaler
}

func (h *Heap[T, TPtr]) GetLessor() interfaces.Lessor3[TPtr] {
	return h.h.Lessor
}

func (h *Heap[T, TPtr]) GetResetter() interfaces.Resetter2[T, TPtr] {
	return h.h.Resetter
}

func (h *Heap[T, TPtr]) Peek() (sk TPtr, ok bool) {
	h.l.Lock()
	defer h.l.Unlock()

	if h.h.Len() > 0 {
		sk = h.h.GetPool().Get()
		h.h.Resetter.ResetWith(sk, h.h.Elements[0])
		ok = true
	}

	return
}

func (h *Heap[T, TPtr]) Add(sk TPtr) (err error) {
	h.Push(sk)
	return
}

func (h *Heap[T, TPtr]) Push(sk TPtr) {
	h.l.Lock()
	defer h.l.Unlock()

	if h.s > 0 {
		panic(
			errors.ErrorWithStackf(
				"attempting to push to a heap that has saved elements",
			),
		)
	}

	heap.Push(&h.h, sk)
}

func (h *Heap[T, TPtr]) PopAndSave() (sk TPtr, ok bool) {
	h.l.Lock()
	defer h.l.Unlock()

	return h.popAndSave()
}

func (h *Heap[T, TPtr]) popAndSave() (sk TPtr, ok bool) {
	// h.h.discardDupes()

	if h.h.Len() == 0 {
		return
	}

	sk = h.h.GetPool().Get()
	e := heap.Pop(&h.h).(TPtr)
	h.h.Resetter.ResetWith(sk, e)
	ok = true
	h.s += 1
	faked := h.h.Elements[:h.h.Len()+h.s]
	faked[h.h.Len()] = e
	h.h.saveLastPopped(sk)

	return
}

func (h *Heap[T, TPtr]) Restore() {
	h.l.Lock()
	defer h.l.Unlock()

	h.restore()
}

func (h *Heap[T, TPtr]) restore() {
	h.h.Elements = h.h.Elements[:h.s]
	h.s = 0
	h.h.GetPool().Put(h.h.lastPopped)
	h.h.lastPopped = nil

	quiter.ReverseSortable(&h.h)
}

func (h *Heap[T, TPtr]) PopError() (sk TPtr, err error) {
	ok := false
	sk, ok = h.Pop()

	if !ok {
		err = errors.MakeErrStopIteration()
	}

	return
}

func (h *Heap[T, TPtr]) Pop() (sk TPtr, ok bool) {
	h.l.Lock()
	defer h.l.Unlock()

	// h.h.discardDupes()

	if h.h.Len() == 0 {
		return
	}

	sk = h.h.GetPool().Get()
	h.h.Resetter.ResetWith(sk, heap.Pop(&h.h).(TPtr))
	ok = true
	h.h.saveLastPopped(sk)

	return
}

func (h *Heap[T, TPtr]) Len() int {
	h.l.Lock()
	defer h.l.Unlock()

	return h.h.Len()
}

func (a *Heap[T, TPtr]) Equals(b *Heap[T, TPtr]) bool {
	a.l.Lock()
	defer a.l.Unlock()

	if a.h.Len() != b.h.Len() {
		return false
	}

	for i, av := range a.h.Elements {
		if b.h.equaler.Equals(b.h.Elements[i], av) {
			return false
		}
	}

	return true
}

func (a *Heap[T, TPtr]) Copy() Heap[T, TPtr] {
	a.l.Lock()
	defer a.l.Unlock()

	return Heap[T, TPtr]{
		h: a.h.Copy(),
	}
}

func (a *Heap[T, TPtr]) Fix() {
	a.l.Lock()
	defer a.l.Unlock()

	heap.Init(&a.h)
}

func (a *Heap[T, TPtr]) Sorted() (b []TPtr) {
	a.l.Lock()
	defer a.l.Unlock()

	b = a.h.Sorted().Elements

	return
}

func (a *Heap[T, TPtr]) Reset() {
	a.h.Elements = make([]TPtr, 0)
	a.h.GetPool().Put(a.h.lastPopped)
	a.h.p = nil
	a.h.lastPopped = nil
}

func (a *Heap[T, TPtr]) ResetWith(b *Heap[T, TPtr]) {
	a.h.equaler = b.h.equaler
	a.h.Lessor = b.h.Lessor
	a.h.Resetter = b.h.Resetter
	a.h.Elements = make([]TPtr, b.Len())

	for i, bv := range b.h.Elements {
		a.h.Elements[i] = bv
	}

	a.h.p = b.h.GetPool()
}
