package id_fmts

import (
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections_ptr"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
)

type tagsReader struct{}

func MakeTagsReader() (f *tagsReader) {
	f = &tagsReader{}

	return f
}

func (f *tagsReader) ReadStringFormat(
	k ids.TagMutableSet,
	rb *catgut.RingBuffer,
) (n int64, err error) {
	flag := collections_ptr.MakeFlagCommasFromExisting(
		collections_ptr.SetterPolicyAppend,
		k,
	)

	var readable catgut.Slice

	if readable, err = rb.PeekUptoAndIncluding('\n'); err != nil && err != io.EOF {
		err = errors.Wrap(err)
		return n, err
	}

	if readable.Len() == 1 {
		return n, err
	}

	tag := strings.TrimSpace(readable.String())

	if err = flag.Set(tag); err != nil {
		if errors.Is(err, ids.ErrEmptyTag) {
			err = nil
		} else {
			err = errors.Wrap(err)
			return n, err
		}
	}

	n = int64(readable.Len())
	rb.AdvanceRead(readable.Len())

	return n, err
}
