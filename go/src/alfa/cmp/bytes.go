package cmp

import "unicode/utf8"

// TODO move to cmp.Result

type Comparable[Self any] interface {
	Len() int
	SliceFrom(int) Self
	DecodeRune() (r rune, width int)
}

func CompareUTF8Bytes[A Comparable[A], B Comparable[B]](
	a A,
	b B,
	partial bool,
) int {
	lenA, lenB := a.Len(), b.Len()

	// TODO remove?
	switch {
	case lenA == 0 && lenB == 0:
		return 0

	case lenA == 0:
		return -1

	case lenB == 0:
		return 1
	}

	for {
		lenA, lenB := a.Len(), b.Len()

		switch {
		case lenA == 0 && lenB == 0:
			return 0

		case lenA == 0:
			if partial && lenB <= lenA {
				return 0
			} else {
				return -1
			}

		case lenB == 0:
			if partial {
				return 0
			} else {
				return 1
			}
		}

		runeA, widthA := a.DecodeRune()
		a = a.SliceFrom(widthA)

		if runeA == utf8.RuneError {
			panic("not a valid utf8 string")
		}

		runeB, widthB := b.DecodeRune()
		b = b.SliceFrom(widthB)

		if runeB == utf8.RuneError {
			panic("not a valid utf8 string")
		}

		if runeA < runeB {
			return -1
		} else if runeA > runeB {
			return 1
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
