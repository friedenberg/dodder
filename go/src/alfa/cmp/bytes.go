package cmp

import "unicode/utf8"

// TODO move to cmp.Result

type Comparable[Self any] interface {
	Len() int
	SliceFrom(int) Self
	DecodeRune() (r rune, width int)
}

func CompareUTF8Bytes[
	LEFT Comparable[LEFT],
	RIGHT Comparable[RIGHT],
](
	left LEFT,
	right RIGHT,
	partial bool,
) Result {
	lenLeft, lenRight := left.Len(), right.Len()

	// TODO remove?
	switch {
	case lenLeft == 0 && lenRight == 0:
		return Equal

	case lenLeft == 0:
		return Less

	case lenRight == 0:
		return Greater
	}

	for {
		lenLeft, lenRight := left.Len(), right.Len()

		switch {
		case lenLeft == 0 && lenRight == 0:
			return Equal

		case lenLeft == 0:
			if partial && lenRight <= lenLeft {
				return Equal
			} else {
				return Less
			}

		case lenRight == 0:
			if partial {
				return Equal
			} else {
				return Greater
			}
		}

		runeLeft, widthLeft := left.DecodeRune()
		left = left.SliceFrom(widthLeft)

		if runeLeft == utf8.RuneError {
			panic("not a valid utf8 string")
		}

		runeRight, widthRight := right.DecodeRune()
		right = right.SliceFrom(widthRight)

		if runeRight == utf8.RuneError {
			panic("not a valid utf8 string")
		}

		if runeLeft < runeRight {
			return Less
		} else if runeLeft > runeRight {
			return Greater
		}
	}
}

type ComparableBytes []byte

func (cb ComparableBytes) Len() int {
	return len(cb)
}

func (cb ComparableBytes) SliceFrom(start int) ComparableBytes {
	return ComparableBytes(cb[start:])
}

func (cb ComparableBytes) DecodeRune() (r rune, width int) {
	r, width = utf8.DecodeRune(cb)
	return r, width
}

type ComparerString string

func (cb ComparerString) Len() int {
	return len(cb)
}

func (cb ComparerString) SliceFrom(start int) ComparerString {
	return ComparerString(cb[start:])
}

func (cb ComparerString) DecodeRune() (r rune, width int) {
	for _, r1 := range cb {
		r = r1
		break
	}

	width = utf8.RuneLen(r)

	return r, width
}
