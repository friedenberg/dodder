package ohio

import (
	"io"
	"unicode/utf8"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func TeeRuneScanner(
	runeReader io.RuneScanner,
	writer io.Writer,
) *teeRuneScanner {
	return &teeRuneScanner{
		runeScanner: runeReader,
		writer:      writer,
	}
}

type teeRuneScanner struct {
	last        rune
	runeScanner io.RuneScanner
	writer      io.Writer
}

func (tee *teeRuneScanner) UnreadRune() (err error) {
	if err = tee.runeScanner.UnreadRune(); err != nil {
		err = errors.Wrap(err)
		return
	}

	tee.last = utf8.RuneError

	return
}

func (tee *teeRuneScanner) ReadRune() (r rune, size int, err error) {
	if r, size, err = tee.runeScanner.ReadRune(); err != nil {
		err = errors.Wrap(err)
		return
	}

	tee.last = r

	return
}
