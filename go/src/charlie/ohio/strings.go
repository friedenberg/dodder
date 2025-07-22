package ohio

import (
	"bytes"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

func WriteKeySpaceValueNewlineString(
	writer io.Writer,
	key, value string,
) (n int, err error) {
	return WriteStrings(writer, key, " ", value, "\n")
}

func WriteKeySpaceValueNewline(
	writer io.Writer,
	key string, value []byte,
) (n int64, err error) {
	var (
		n1           int64
		stringReader *strings.Reader
		byteReader   *bytes.Reader
	)

	stringReader = strings.NewReader(key)
	n1, err = stringReader.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	stringReader = strings.NewReader(" ")

	n1, err = stringReader.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	byteReader = bytes.NewReader(value)

	n1, err = byteReader.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	stringReader = strings.NewReader("\n")

	n1, err = stringReader.WriteTo(writer)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func WriteStrings(
	w io.Writer,
	ss ...string,
) (n int, err error) {
	for _, s := range ss {
		var n1 int

		n1, err = io.WriteString(w, s)
		n += n1

		if err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}
