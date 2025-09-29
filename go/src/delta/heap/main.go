package heap

import (
	pkg_heap "container/heap"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

// TODO rewrite with quiter.SortComparer
type Heap[T Element, TPtr ElementPtr[T]] struct {
	lock       sync.Mutex
	private    heapPrivate[T, TPtr]
	savedIndex int
}

func (heap *Heap[T, TPtr]) GetCollection() interfaces.Collection[TPtr] {
	return heap
}

func (heap *Heap[T, TPtr]) Any() TPtr {
	element, _ := heap.Peek()
	return element
}

func (heap *Heap[T, TPtr]) All() interfaces.Seq[TPtr] {
	return func(yield func(TPtr) bool) {
		heap.lock.Lock()
		defer heap.lock.Unlock()
		defer heap.restore()

		for {
			element, ok := heap.popAndSave()

			if !ok {
				return
			}

			if !yield(element) {
				break
			}
		}
	}
}

func (heap *Heap[T, TPtr]) SetPool(v interfaces.Pool[T, TPtr]) {
	heap.private.pool = v
}

func (heap *Heap[T, TPtr]) GetEqualer() interfaces.Equaler[TPtr] {
	return heap.private.equaler
}

func (heap *Heap[T, TPtr]) GetLessor() interfaces.Lessor3[TPtr] {
	return heap.private.Lessor
}

func (heap *Heap[T, TPtr]) GetResetter() interfaces.Resetter2[T, TPtr] {
	return heap.private.Resetter
}

func (heap *Heap[T, TPtr]) Peek() (element TPtr, ok bool) {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	if heap.private.Len() > 0 {
		element = heap.private.Elements[0]
		ok = true
	}

	return element, ok
}

func (heap *Heap[T, TPtr]) Add(element TPtr) (err error) {
	heap.Push(element)
	return err
}

func (heap *Heap[T, TPtr]) Push(element TPtr) {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	if heap.savedIndex > 0 {
		panic(
			errors.ErrorWithStackf(
				"attempting to push to a heap that has saved elements",
			),
		)
	}

	pkg_heap.Push(&heap.private, element)
}

func (heap *Heap[T, TPtr]) PopAndSave() (element TPtr, ok bool) {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	return heap.popAndSave()
}

func (heap *Heap[T, TPtr]) popAndSave() (element TPtr, ok bool) {
	// h.h.discardDupes()

	if heap.private.Len() == 0 {
		return element, ok
	}

	element = heap.private.GetPool().Get()
	e := pkg_heap.Pop(&heap.private).(TPtr)
	heap.private.Resetter.ResetWith(element, e)
	ok = true
	heap.savedIndex += 1
	faked := heap.private.Elements[:heap.private.Len()+heap.savedIndex]
	faked[heap.private.Len()] = e
	heap.private.saveLastPopped(element)

	return element, ok
}

func (heap *Heap[T, TPtr]) Restore() {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	heap.restore()
}

func (heap *Heap[T, TPtr]) restore() {
	heap.private.Elements = heap.private.Elements[:heap.savedIndex]
	heap.savedIndex = 0
	heap.private.GetPool().Put(heap.private.lastPopped)
	heap.private.lastPopped = nil

	quiter.ReverseSortable(&heap.private)
}

func (heap *Heap[T, TPtr]) PopOrErrStopIteration() (element TPtr, err error) {
	ok := false
	element, ok = heap.Pop()

	if !ok {
		err = errors.MakeErrStopIteration()
	}

	return element, err
}

func (heap *Heap[T, TPtr]) PopError() (element TPtr, err error) {
	element, _ = heap.Pop()
	return element, err
}

func (heap *Heap[T, TPtr]) Pop() (element TPtr, ok bool) {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	// h.h.discardDupes()

	if heap.private.Len() == 0 {
		return element, ok
	}

	element = heap.private.GetPool().Get()
	heap.private.Resetter.ResetWith(element, pkg_heap.Pop(&heap.private).(TPtr))
	ok = true
	heap.private.saveLastPopped(element)

	return element, ok
}

func (heap *Heap[T, TPtr]) Len() int {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	return heap.private.Len()
}

func (heap *Heap[T, TPtr]) Equals(b *Heap[T, TPtr]) bool {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	if heap.private.Len() != b.private.Len() {
		return false
	}

	for i, av := range heap.private.Elements {
		if b.private.equaler.Equals(b.private.Elements[i], av) {
			return false
		}
	}

	return true
}

func (heap *Heap[T, TPtr]) Copy() Heap[T, TPtr] {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	return Heap[T, TPtr]{
		private: heap.private.Copy(),
	}
}

func (heap *Heap[T, TPtr]) Fix() {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	pkg_heap.Init(&heap.private)
}

func (heap *Heap[T, TPtr]) Sorted() (heapCopy []TPtr) {
	heap.lock.Lock()
	defer heap.lock.Unlock()

	heapCopy = heap.private.Sorted().Elements

	return heapCopy
}

func (heap *Heap[T, TPtr]) Reset() {
	heap.private.Elements = make([]TPtr, 0)
	heap.private.GetPool().Put(heap.private.lastPopped)
	heap.private.pool = nil
	heap.private.lastPopped = nil
}

func (heap *Heap[T, TPtr]) ResetWith(b *Heap[T, TPtr]) {
	heap.private.equaler = b.private.equaler
	heap.private.Lessor = b.private.Lessor
	heap.private.Resetter = b.private.Resetter
	heap.private.Elements = make([]TPtr, b.Len())

	copy(heap.private.Elements, b.private.Elements)

	heap.private.pool = b.private.GetPool()
}
