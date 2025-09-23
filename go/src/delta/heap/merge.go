package heap

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// TODO make compatible with iterators
func MergeStream[ELEMENT Element, ELEMENT_PTR ElementPtr[ELEMENT]](
	heap *Heap[ELEMENT, ELEMENT_PTR],
	read func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
) (err error) {
	if err = MergeStreamPreferringHeap(
		heap,
		read,
		write,
		heap.h.equaler,
		heap.h.Lessor,
		heap.h.Resetter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO make compatible with iterators
func MergeStreamPreferringHeap[T Element, TPtr ElementPtr[T]](
	heap *Heap[T, TPtr],
	read func() (TPtr, error),
	write interfaces.FuncIter[TPtr],
	equaler interfaces.Equaler[TPtr],
	lessor interfaces.Lessor3[TPtr],
	resetter interfaces.Resetter2[T, TPtr],
) (err error) {
	defer func() {
		heap.restore()
	}()

	oldWrite := write

	var last TPtr

	write = func(e TPtr) (err error) {
		if last == nil {
			var t T
			last = &t
		} else if equaler.Equals(e, last) {
			return err
		} else if lessor.Less(e, last) {
			err = errors.ErrorWithStackf(
				"last is greater than current! last:\n%v\ncurrent: %v",
				last,
				e,
			)

			return err
		}

		resetter.ResetWith(last, e)

		return oldWrite(e)
	}

	for {
		var element TPtr

		if element, err = read(); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return err
			}
		}

	LOOP:
		for {
			peeked, ok := heap.Peek()

			switch {
			case !ok:
				break LOOP

			case equaler.Equals(peeked, element):
				element = peeked
				heap.Pop()
				continue LOOP

			case lessor.Less(element, peeked):
				break LOOP

			default:
			}

			popped, _ := heap.popAndSave()

			if !equaler.Equals(peeked, popped) {
				err = errors.ErrorWithStackf(
					"popped not equal to peeked: %s != %s",
					popped,
					peeked,
				)

				return err
			}

			if popped == nil {
				break
			}

			if err = write(popped); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
				}

				return err
			}
		}

		if err = write(element); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}
	}

	for {
		popped, ok := heap.popAndSave()

		if !ok {
			break
		}

		if err = write(popped); err != nil {
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
