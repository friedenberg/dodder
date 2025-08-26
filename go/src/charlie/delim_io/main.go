package delim_io

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var poolReader = pool.MakeWithResetable[reader]()

func PutReader(dr *reader) {
	poolReader.Put(dr)
}

// Not safe for parallel use
type Reader interface {
	io.Reader
	N() int64
	Segments() int64
	IsEOF() bool
	ResetWith(dr reader)
	Reset()
	ReadOneString() (str string, err error)
	ReadOneKeyValue(sep string) (key, val string, err error)
}

type reader struct {
	delim byte
	bufio.Reader
	n         int64
	lastReadN int
	segments  int64
	eof       bool
}

func Make(
	delim byte,
	reader io.Reader,
) (delimReader *reader) {
	delimReader = poolReader.Get()
	delimReader.Reader.Reset(reader)
	delimReader.delim = delim

	return
}

func (delimReader *reader) N() int64 {
	return delimReader.n
}

func (delimReader *reader) Segments() int64 {
	return delimReader.segments
}

func (delimReader *reader) IsEOF() bool {
	return delimReader.eof
}

func (delimReader *reader) ResetWith(dr reader) {
	delimReader.Reader.Reset(nil)
	delimReader.delim = dr.delim
}

func (delimReader *reader) Reset() {
	delimReader.Reader.Reset(nil)
	delimReader.n = 0
	delimReader.lastReadN = 0
	delimReader.segments = 0
	delimReader.eof = false
}

func (delimReader *reader) ReadOneBytes() (str []byte, err error) {
	if delimReader.eof {
		err = io.EOF
		return
	}

	var rawLine []byte

	rawLine, err = delimReader.Reader.ReadSlice(delimReader.delim)
	delimReader.lastReadN = len(rawLine)
	delimReader.n += int64(delimReader.lastReadN)

	if err != nil && err != io.EOF {
		err = errors.Wrap(err)
		return
	}

	if err == io.EOF {
		delimReader.eof = true
	}

	str = bytes.TrimSuffix(rawLine, []byte{delimReader.delim})

	delimReader.segments++

	return
}

// Not safe for parallel use
func (delimReader *reader) ReadOneString() (str string, err error) {
	if delimReader.eof {
		err = io.EOF
		return
	}

	var rawLine string

	rawLine, err = delimReader.Reader.ReadString(delimReader.delim)
	delimReader.lastReadN = len(rawLine)
	delimReader.n += int64(delimReader.lastReadN)

	if err != nil && err != io.EOF {
		err = errors.Wrap(err)
		return
	}

	if err == io.EOF {
		delimReader.eof = true
	}

	str = strings.TrimSuffix(rawLine, string([]byte{delimReader.delim}))

	delimReader.segments++

	return
}

// Not safe for parallel use
func (delimReader *reader) ReadOneKeyValue(
	sep string,
) (key, val string, err error) {
	if delimReader.eof {
		err = io.EOF
		return
	}

	str, err := delimReader.ReadOneString()
	if err != nil {
		if err == io.EOF {
			delimReader.eof = true
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	loc := strings.Index(str, sep)

	if loc == -1 {
		err = errors.ErrorWithStackf(
			"expected at least one %q, but found none: %q",
			sep,
			str,
		)
		return
	}

	key = str[:loc]
	val = str[loc+1:]

	return
}

func (delimReader *reader) ReadOneKeyValueBytes(
	sep byte,
) (key, val []byte, err error) {
	if delimReader.eof {
		err = io.EOF
		return
	}

	str, err := delimReader.ReadOneBytes()
	if err != nil {
		if err == io.EOF {
			delimReader.eof = true
		} else {
			err = errors.Wrap(err)
		}

		return
	}

	loc := bytes.IndexByte(str, sep)

	if loc == -1 {
		err = errors.ErrorWithStackf(
			"expected at least one %q, but found none: %q",
			sep,
			str,
		)
		return
	}

	key = str[:loc]
	val = str[loc+1:]

	return
}
