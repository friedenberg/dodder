package collections_slice

import "unicode/utf8"

type (
	Byte struct {
		Slice[byte]
	}

	Rune struct {
		Slice[rune]
	}
)

func (slice Byte) ReadRune() (char rune, width int, err error) {
	return
}

func (slice Byte) Shift(amount int) Byte {
	slice.Slice = slice.Slice[amount:]
	return slice
}

func (slice Rune) ReadRune() (char rune, width int, err error) {
	for _, char = range slice.Slice {
		break
	}

	width = utf8.RuneLen(char)

	return
}

func (slice Rune) Shift(amount int) Rune {
	slice.Slice = slice.Slice[amount:]
	return slice
}
