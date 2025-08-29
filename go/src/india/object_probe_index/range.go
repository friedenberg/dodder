package object_probe_index

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Range struct {
	Offset, ContentLength int64
}

func (raynge *Range) Size() int {
	return 8 * 2
}

func (raynge Range) IsEmpty() bool {
	return raynge.Offset == 0 && raynge.ContentLength == 0
}

func (raynge Range) String() string {
	return fmt.Sprintf("%03d+%03d", raynge.Offset, raynge.ContentLength)
}

func (raynge *Range) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int
	n1, raynge.Offset, err = ohio.ReadFixedInt64(r)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, raynge.ContentLength, err = ohio.ReadFixedInt64(r)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (raynge *Range) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int
	n1, err = ohio.WriteFixedInt64(w, raynge.Offset)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = ohio.WriteFixedInt64(w, raynge.ContentLength)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}
