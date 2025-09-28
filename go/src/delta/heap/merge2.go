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
	if err = MergeStreamPreferringAdditions(
		heapAdditions.PopError,
		readHistory,
		write,
		cmp.MakeComparerFromEqualerAndLessor3EqualFirst(
			heapAdditions.private.equaler,
			heapAdditions.private.Lessor,
		),
		heapAdditions.private.Resetter,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func MergeStreamPreferringAdditions[
	ELEMENT Element,
	ELEMENT_PTR ElementPtr[ELEMENT],
](
	readLeft, readRight func() (ELEMENT_PTR, error),
	write interfaces.FuncIter[ELEMENT_PTR],
	funcCmp func(ELEMENT_PTR, ELEMENT_PTR) cmp.Result,
	resetter interfaces.Resetter2[ELEMENT, ELEMENT_PTR],
) (err error) {
	oldWrite := write

	var lastElement ELEMENT_PTR

	write = func(element ELEMENT_PTR) (err error) {
		if lastElement == nil {
			var elementValue ELEMENT
			lastElement = &elementValue
		} else {
			switch funcCmp(lastElement, element) {
			case cmp.Less:
				break

			case cmp.Equal:
				return err

			case cmp.Greater:
				err = errors.ErrorWithStackf(
					"last is greater than current! last:\n%v\ncurrent: %v",
					lastElement,
					element,
				)

				return err
			}
		}

		resetter.ResetWith(lastElement, element)

		// TODO figure out why writing lastElement fails
		return oldWrite(element)
	}

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
			if right.carry, err = right.read(); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
					remaining = left
					break
				} else {
					err = errors.Wrap(err)
					return err
				}
			}

			if right.carry == nil {
				remaining = left
				break
			}
		}

		if left.carry == nil {
			if left.carry, err = left.read(); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
					remaining = right
					break
				} else {
					err = errors.Wrap(err)
					return err
				}
			}

			if left.carry == nil {
				remaining = right
				break
			}
		}

		switch funcCmp(left.carry, right.carry) {
		case cmp.Less, cmp.Equal:
			if err = write(left.carry); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
				}

				return err
			}

			left.carry = nil

		case cmp.Greater:
			if err = write(right.carry); err != nil {
				if errors.IsStopIteration(err) {
					err = nil
				} else {
					err = errors.Wrap(err)
				}

				return err
			}

			right.carry = nil
		}
	}

	if remaining.carry != nil {
		if err = write(remaining.carry); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}
	}

	for {
		var element ELEMENT_PTR

		if element, err = remaining.read(); err != nil {
			if errors.IsStopIteration(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}

		if element == nil {
			return err
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

	return err
}
