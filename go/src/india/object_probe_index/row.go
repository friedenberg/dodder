package object_probe_index

import (
	"bytes"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

func (page *page) writeIntoRow(
	row *row,
	reader io.Reader,
) (n int64, err error) {
	if n, err = row.ReadFrom(reader, page.hashType); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	if err = markl.MakeErrLength(int64(page.rowWidth), n); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (page *page) readFromRow(
	row *row,
	writer io.Writer,
) (n int64, err error) {
	if n, err = row.WriteTo(writer); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	if err = markl.MakeErrLength(int64(page.rowWidth), n); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type row struct {
	Digest markl.Id
	Loc
}

func (row *row) IsEmpty() bool {
	return row.Loc.IsEmpty() && row.Digest.IsNull()
}

func (row *row) String() string {
	return fmt.Sprintf(
		"%s %s",
		&row.Loc,
		row.Digest.String(),
	)
}

func (row *row) ReadFrom(
	reader io.Reader,
	hashType markl.HashType,
) (n int64, err error) {
	var n1 int
	var n2 int64

	n1, err = markl.ReadFrom(reader, &row.Digest, hashType)
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	n2, err = row.Loc.ReadFrom(reader)
	n += int64(n2)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (row *row) WriteTo(writer io.Writer) (n int64, err error) {
	var n1 int
	var n2 int64

	n1, err = writer.Write(row.Digest.GetBytes())
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
	return markl.Equals(&a.Digest, &b.Digest) &&
		a.Loc.Page == b.Loc.Page &&
		a.Loc.Offset == b.Loc.Offset &&
		a.Loc.ContentLength == b.Loc.ContentLength
}

type rowEqualerDigestOnly struct{}

func (rowEqualerDigestOnly) Equals(a, b *row) bool {
	return markl.Equals(&a.Digest, &b.Digest)
}

type rowResetter struct{}

func (rowResetter) Reset(a *row) {
	a.Digest.Reset()
	a.Page = 0
	a.Offset = 0
	a.ContentLength = 0
}

func (rowResetter) ResetWith(a, b *row) {
	a.Digest.ResetWith(b.Digest)
	a.Page = b.Page
	a.Offset = b.Offset
	a.ContentLength = b.ContentLength
}

type rowLessor struct{}

func (rowLessor) Less(a, b *row) bool {
	cmp := bytes.Compare(a.Digest.GetBytes(), b.Digest.GetBytes())

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
