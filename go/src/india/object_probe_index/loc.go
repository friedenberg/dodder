package object_probe_index

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Loc struct {
	Page uint8
	Range
}

func (l Loc) IsEmpty() bool {
	return l.Page == 0 && l.Offset == 0 && l.ContentLength == 0
}

func (l Loc) String() string {
	return fmt.Sprintf("%02d@%s", l.Page, l.Range)
}

func (l *Loc) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int
	l.Page, n1, err = ohio.ReadFixedUint8(r)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	var n2 int64
	n2, err = l.Range.ReadFrom(r)
	n += n2

	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	return
}

func (l *Loc) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteUint8(w, l.Page)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	var n2 int64
	n2, err = l.Range.WriteTo(w)
	n += n2

	if err != nil {
		err = errors.WrapExcept(err, io.EOF)
		return
	}

	return
}
