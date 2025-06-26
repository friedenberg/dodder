package delim_io

import (
	"bufio"
	"bytes"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var _pool interfaces.Pool[reader, *reader]

func init() {
	_pool = pool.MakePoolWithReset[reader]()
}

func PutReader(dr *reader) {
	_pool.Put(dr)
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
	*bufio.Reader
	n         int64
	lastReadN int
	segments  int64
	eof       bool
}

func Make(
	delim byte,
	r io.Reader,
) (dr *reader) {
	dr = _pool.Get()
	dr.Reader.Reset(r)
	dr.delim = delim

	return
}

func (lr *reader) N() int64 {
	return lr.n
}

func (lr *reader) Segments() int64 {
	return lr.segments
}

func (lr *reader) IsEOF() bool {
	return lr.eof
}

func (lr *reader) ResetWith(dr reader) {
	lr.Reader.Reset(nil)
	lr.delim = dr.delim
}

func (lr *reader) Reset() {
	if lr.Reader == nil {
		lr.Reader = bufio.NewReader(nil)
	} else {
		lr.Reader.Reset(nil)
	}

	lr.n = 0
	lr.lastReadN = 0
	lr.segments = 0
	lr.eof = false
}

func (lr *reader) ReadOneBytes() (str []byte, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	var rawLine []byte

	rawLine, err = lr.Reader.ReadSlice(lr.delim)
	lr.lastReadN = len(rawLine)
	lr.n += int64(lr.lastReadN)

	if err != nil && err != io.EOF {
		err = errors.Wrap(err)
		return
	}

	if err == io.EOF {
		lr.eof = true
	}

	str = bytes.TrimSuffix(rawLine, []byte{lr.delim})

	lr.segments++

	return
}

// Not safe for parallel use
func (lr *reader) ReadOneString() (str string, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	var rawLine string

	rawLine, err = lr.Reader.ReadString(lr.delim)
	lr.lastReadN = len(rawLine)
	lr.n += int64(lr.lastReadN)

	if err != nil && err != io.EOF {
		err = errors.Wrap(err)
		return
	}

	if err == io.EOF {
		lr.eof = true
	}

	str = strings.TrimSuffix(rawLine, string([]byte{lr.delim}))

	lr.segments++

	return
}

// Not safe for parallel use
func (lr *reader) ReadOneKeyValue(
	sep string,
) (key, val string, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	str, err := lr.ReadOneString()
	if err != nil {
		if err == io.EOF {
			lr.eof = true
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

func (lr *reader) ReadOneKeyValueBytes(
	sep byte,
) (key, val []byte, err error) {
	if lr.eof {
		err = io.EOF
		return
	}

	str, err := lr.ReadOneBytes()
	if err != nil {
		if err == io.EOF {
			lr.eof = true
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
