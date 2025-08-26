package ohio

import (
	"bytes"
	"io"
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO add tests
func DebugRuneScanner(
	runeReader io.RuneScanner,
) *teeRuneScanner {
	return &teeRuneScanner{
		runeScanner: runeReader,
	}
}

type teeRuneScanner struct {
	lastWidth   int
	runeScanner io.RuneScanner
	buffer      bytes.Buffer
}

func (tee *teeRuneScanner) UnreadRune() (err error) {
	if err = tee.runeScanner.UnreadRune(); err != nil {
		err = errors.Wrap(err)
		return
	}

	tee.buffer.Truncate(tee.buffer.Len() - tee.lastWidth)
	tee.lastWidth = utf8.RuneError

	return
}

func (tee *teeRuneScanner) ReadRune() (r rune, size int, err error) {
	if r, size, err = tee.runeScanner.ReadRune(); err != nil {
		if err == io.EOF {
			return
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	tee.lastWidth = size

	if _, err = tee.buffer.WriteRune(r); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (tee *teeRuneScanner) GetBytes() []byte {
	return tee.buffer.Bytes()
}
