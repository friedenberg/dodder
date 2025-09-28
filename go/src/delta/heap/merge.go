package heap

import (
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
	seq := MergeStreamPreferringAdditions(
		heapAdditions.PopError,
		readHistory,
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
	readLeft, readRight func() (ELEMENT_PTR, error),
	funcCmp func(ELEMENT_PTR, ELEMENT_PTR) cmp.Result,
	resetter interfaces.Resetter2[ELEMENT, ELEMENT_PTR],
) interfaces.SeqError[ELEMENT_PTR] {
	return func(yield func(ELEMENT_PTR, error) bool) {
		type readWithElement struct {
			read  func() (ELEMENT_PTR, error)
			carry ELEMENT_PTR
		}

		left := readWithElement{
			read: readLeft,
		}

		right := readWithElement{
			read: readRight,
		}

		var remaining readWithElement

		for {
			if right.carry == nil {
				var err error
				if right.carry, err = right.read(); err != nil {
					if errors.IsStopIteration(err) {
						remaining = left
						goto EMIT_REMAINING
					} else {
						yield(nil, err)
						return
					}
				}

				if right.carry == nil {
					remaining = left
					break
				}
			}

			if left.carry == nil {
				var err error
				if left.carry, err = left.read(); err != nil {
					if errors.IsStopIteration(err) {
						remaining = right
						goto EMIT_REMAINING
					} else {
						yield(nil, err)
						return
					}
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

			if element, err = remaining.read(); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
					yield(nil, err)
				}

				return
			}

			if element == nil {
				return
			}

			if !yield(element, nil) {
				return
			}
		}
	}
}
