package object_probe_index

import (
	"bytes"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

type row struct {
	BlobId sha.Sha
	Loc
}

func writeIntoRow(
	row *row,
	reader io.Reader,
	rowWidth int,
) (n int64, err error) {
	if n, err = row.ReadFrom(reader); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	if err = merkle.MakeErrLength(int64(rowWidth), n); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func readFromRow(
	row *row,
	writer io.Writer,
	rowWidth int,
) (n int64, err error) {
	if n, err = row.WriteTo(writer); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	if err = merkle.MakeErrLength(int64(rowWidth), n); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (row *row) IsEmpty() bool {
	return row.Loc.IsEmpty() && row.BlobId.IsNull()
}

func (row *row) String() string {
	return fmt.Sprintf(
		"%s %s",
		&row.Loc,
		row.BlobId.String(),
	)
}

func (row *row) ReadFrom(reader io.Reader) (n int64, err error) {
	var n1 int64

	n1, err = row.BlobId.ReadFrom(reader)
	n += n1

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	n1, err = row.Loc.ReadFrom(reader)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (row *row) WriteTo(writer io.Writer) (n int64, err error) {
	var n1 int
	var n2 int64

	n, err = row.BlobId.WriteTo(writer)
	n += int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n2, err = row.Loc.WriteTo(writer)
	n += int64(n2)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type rowEqualerComplete struct{}

func (rowEqualerComplete) Equals(a, b *row) bool {
	return merkle.Equals(&a.BlobId, &b.BlobId) &&
		a.Loc.Page == b.Loc.Page &&
		a.Loc.Offset == b.Loc.Offset &&
		a.Loc.ContentLength == b.Loc.ContentLength
}

type rowEqualerShaOnly struct{}

func (rowEqualerShaOnly) Equals(a, b *row) bool {
	return merkle.Equals(&a.BlobId, &b.BlobId)
}

type rowResetter struct{}

func (rowResetter) Reset(a *row) {
	a.BlobId.Reset()
	a.Page = 0
	a.Offset = 0
	a.ContentLength = 0
}

func (rowResetter) ResetWith(a, b *row) {
	a.BlobId.ResetWith(&b.BlobId)
	a.Page = b.Page
	a.Offset = b.Offset
	a.ContentLength = b.ContentLength
}

type rowLessor struct{}

func (rowLessor) Less(a, b *row) bool {
	cmp := bytes.Compare(a.BlobId.GetBytes(), b.BlobId.GetBytes())

	if cmp != 0 {
		return cmp == -1
	}

	if a.Page != b.Page {
		return a.Page < b.Page
	}

	if a.Offset != b.Offset {
		return a.Offset < b.Offset
	}

	return a.ContentLength < b.ContentLength
}
