package object_probe_index

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Loc struct {
	Page uint8
	Cursor
}

func (loc Loc) IsEmpty() bool {
	return loc.Page == 0 && loc.Offset == 0 && loc.ContentLength == 0
}

func (loc Loc) String() string {
	return fmt.Sprintf("%02d@%s", loc.Page, loc.Cursor)
}

func (loc *Loc) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int
	loc.Page, n1, err = ohio.ReadFixedUint8(r)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	var n2 int64
	n2, err = loc.Cursor.ReadFrom(r)
	n += n2

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (loc *Loc) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteUint8(w, loc.Page)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	var n2 int64
	n2, err = loc.Cursor.WriteTo(w)
	n += n2

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}
