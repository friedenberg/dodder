package heap

import (
	"iter"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
)

func MergeHeap[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	heapAdditions *Heap[ELEMENT, ELEMENT_PTR],
	readHistory func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
) (err error) {
	pullLeft, stopLeft := iter.Pull2(heapAdditions.AllError())
	defer stopLeft()

	pullRight := func() (ELEMENT_PTR, error, bool) {
		element, err := readHistory()
		return element, err, !errors.IsStopIteration(err)
	}

	seq := MergeStreamPreferringAdditions(
		pullLeft,
		pullRight,
		cmp.MakeComparerFromEqualerAndLessor3EqualFirst(
			heapAdditions.private.equaler,
			heapAdditions.private.Lessor,
		),
		heapAdditions.private.Resetter,
	)

	for element, errIter := range seq {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return err
		}

		if err = write(element); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func MergeStreamPreferringAdditions[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	pullLeft, pullRight interfaces.Pull2[ELEMENT_PTR, error],
	funcCmp func(ELEMENT_PTR, ELEMENT_PTR) cmp.Result,
	resetter interfaces.Resetter2[ELEMENT, ELEMENT_PTR],
) interfaces.SeqError[ELEMENT_PTR] {
	return func(yield func(ELEMENT_PTR, error) bool) {
		type readWithElement struct {
			read  func() (ELEMENT_PTR, error, bool)
			carry ELEMENT_PTR
		}

		left := readWithElement{
			read: pullLeft,
		}

		right := readWithElement{
			read: pullRight,
		}

		var remaining readWithElement

		for {
			if right.carry == nil {
				var err error
				var ok bool
				if right.carry, err, ok = right.read(); !ok {
					remaining = left
					goto EMIT_REMAINING
				} else if err != nil {
					yield(nil, err)
					return
				}

				if right.carry == nil {
					remaining = left
					break
				}
			}

			if left.carry == nil {
				var err error
				var ok bool
				if left.carry, err, ok = left.read(); !ok {
					remaining = right
					goto EMIT_REMAINING
				} else if err != nil {
					yield(nil, err)
					return
				}

				if left.carry == nil {
					remaining = right
					break
				}
			}

			switch funcCmp(left.carry, right.carry) {
			case cmp.Equal:
				right.carry = nil
				fallthrough

			case cmp.Less:
				if !yield(left.carry, nil) {
					return
				}

				left.carry = nil

			case cmp.Greater:
				if !yield(right.carry, nil) {
					return
				}

				right.carry = nil
			}
		}

	EMIT_REMAINING:

		if remaining.carry != nil {
			if !yield(remaining.carry, nil) {
				return
			}
		}

		for {
			var element ELEMENT_PTR
			var err error
			var ok bool

			if element, err, ok = remaining.read(); !ok {
				return
			} else if err != nil {
				err = errors.Wrap(err)
				yield(nil, err)
				return
			}

			if !yield(element, nil) {
				return
			}
		}
	}
}
