package heap

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func MergeHeap[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	heap *Heap[ELEMENT, ELEMENT_PTR],
	read func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
) (err error) {
	if err = MergeStreamPreferringHeap(
		heap,
		read,
		write,
		heap.private.equaler,
		heap.private.Lessor,
		heap.private.Resetter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func MergeHeapAndRestore[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	heap *Heap[ELEMENT, ELEMENT_PTR],
	read func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
) (err error) {
	if err = MergeStreamPreferringHeapAndRestore(
		heap,
		read,
		write,
		heap.private.equaler,
		heap.private.Lessor,
		heap.private.Resetter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

type stream[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
] interface {
	Peek() (ELEMENT_PTR, bool)
	Pop() (ELEMENT_PTR, bool)
}

func MergeStreamPreferringHeap[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	heap stream[ELEMENT, ELEMENT_PTR],
	read func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
	equaler interfaces.Equaler[ELEMENT_PTR],
	lessor interfaces.Lessor3[ELEMENT_PTR],
	resetter interfaces.Resetter2[ELEMENT, ELEMENT_PTR],
) (err error) {
	oldWrite := write

	var lastElement ELEMENT_PTR

	write = func(e ELEMENT_PTR) (err error) {
		if lastElement == nil {
			var t ELEMENT
			lastElement = &t
		} else if equaler.Equals(e, lastElement) {
			return err
		} else if lessor.Less(e, lastElement) {
			err = errors.ErrorWithStackf(
				"last is greater than current! last:\n%v\ncurrent: %v",
				lastElement,
				e,
			)

			return err
		}

		resetter.ResetWith(lastElement, e)

		return oldWrite(e)
	}

	// write all the mingled elements

	for {
		var element ELEMENT_PTR

		if element, err = read(); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
				break
			} else {
				err = errors.Wrap(err)
				return err
			}
		}

	READ_FROM_HEAP:
		for {
			peekedElement, hasElement := heap.Peek()

			switch {
			case !hasElement:
				break READ_FROM_HEAP

			case equaler.Equals(peekedElement, element):
				element = peekedElement
				heap.Pop()
				continue READ_FROM_HEAP

			case lessor.Less(element, peekedElement):
				break READ_FROM_HEAP

			default:
			}

			if err = write(peekedElement); err != nil {
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

	// write all the new elements

	for {
		popped, ok := heap.Pop()

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

type streamWithRestore[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
] interface {
	Peek() (ELEMENT_PTR, bool)
	Pop() (ELEMENT_PTR, bool)

	restore()
	popAndSave() (ELEMENT_PTR, bool)
}

func MergeStreamPreferringHeapAndRestore[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	heap streamWithRestore[ELEMENT, ELEMENT_PTR],
	read func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
	equaler interfaces.Equaler[ELEMENT_PTR],
	lessor interfaces.Lessor3[ELEMENT_PTR],
	resetter interfaces.Resetter2[ELEMENT, ELEMENT_PTR],
) (err error) {
	defer func() {
		heap.restore()
	}()

	oldWrite := write

	var lastElement ELEMENT_PTR

	write = func(e ELEMENT_PTR) (err error) {
		if lastElement == nil {
			var t ELEMENT
			lastElement = &t
		} else if equaler.Equals(e, lastElement) {
			return err
		} else if lessor.Less(e, lastElement) {
			err = errors.ErrorWithStackf(
				"last is greater than current! last:\n%v\ncurrent: %v",
				lastElement,
				e,
			)

			return err
		}

		resetter.ResetWith(lastElement, e)

		return oldWrite(e)
	}

	for {
		var element ELEMENT_PTR

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
			peekedElement, hasElement := heap.Peek()

			switch {
			case !hasElement:
				break LOOP

			case equaler.Equals(peekedElement, element):
				element = peekedElement
				heap.Pop()
				continue LOOP

			case lessor.Less(element, peekedElement):
				break LOOP

			default:
			}

			poppedElement, _ := heap.popAndSave()

			if !equaler.Equals(peekedElement, poppedElement) {
				err = errors.ErrorWithStackf(
					"popped not equal to peeked: %s != %s",
					poppedElement,
					peekedElement,
				)

				return err
			}

			if poppedElement == nil {
				break
			}

			if err = write(poppedElement); err != nil {
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
