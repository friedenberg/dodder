package heap

import (
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Make[T Element, TPtr ElementPtr[T]](
	equaler interfaces.Equaler[TPtr],
	lessor interfaces.Lessor3[TPtr],
	resetter interfaces.Resetter2[T, TPtr],
) *Heap[T, TPtr] {
	return &Heap[T, TPtr]{
		h: heapPrivate[T, TPtr]{
			equaler:  equaler,
			Lessor:   lessor,
			Resetter: resetter,
			Elements: make([]TPtr, 0),
		},
	}
}

func MakeHeapFromSliceUnsorted[T Element, TPtr ElementPtr[T]](
	equaler interfaces.Equaler[TPtr],
	lessor interfaces.Lessor3[TPtr],
	resetter interfaces.Resetter2[T, TPtr],
	s []TPtr,
) *Heap[T, TPtr] {
	h := heapPrivate[T, TPtr]{
		Lessor:   lessor,
		Resetter: resetter,
		Elements: s,
		equaler:  equaler,
	}

	sort.Sort(h)

	return &Heap[T, TPtr]{
		h: h,
	}
}

func MakeHeapFromSlice[T Element, TPtr ElementPtr[T]](
	equaler interfaces.Equaler[TPtr],
	lessor interfaces.Lessor3[TPtr],
	resetter interfaces.Resetter2[T, TPtr],
	s []TPtr,
) *Heap[T, TPtr] {
	h := heapPrivate[T, TPtr]{
		Lessor:   lessor,
		Resetter: resetter,
		Elements: s,
		equaler:  equaler,
	}

	return &Heap[T, TPtr]{
		h: h,
	}
}
