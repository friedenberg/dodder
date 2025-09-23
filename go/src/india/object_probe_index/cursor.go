package object_probe_index

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Cursor struct {
	Offset, ContentLength int64
}

func (cursor *Cursor) Size() int {
	return 8 * 2
}

func (cursor Cursor) IsEmpty() bool {
	return cursor.Offset == 0 && cursor.ContentLength == 0
}

func (cursor Cursor) String() string {
	return fmt.Sprintf("%03d+%03d", cursor.Offset, cursor.ContentLength)
}

func (cursor *Cursor) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int
	n1, cursor.Offset, err = ohio.ReadFixedInt64(r)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, cursor.ContentLength, err = ohio.ReadFixedInt64(r)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return n, err
	}

	return n, err
}

func (cursor *Cursor) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteFixedInt64(w, cursor.Offset)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n1, err = ohio.WriteFixedInt64(w, cursor.ContentLength)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return n, err
	}

	return n, err
}
