package object_probe_index

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func (page *page) compareRowForFlush(left, right *row) cmp.Result {
	if (rowEqualerDigestOnly{}).Equals(left, right) {
		return cmp.Equal
	}

	if cmp := cmp.Bytes(left.Digest.GetBytes(), right.Digest.GetBytes()); !cmp.Equal() {
		return cmp
	}

	if cmp := cmp.Ordered(left.Page, right.Page); !cmp.Equal() {
		return cmp
	}

	if cmp := cmp.Ordered(left.Offset, right.Offset); !cmp.Equal() {
		return cmp
	}

	return cmp.Ordered(left.ContentLength, right.ContentLength)
}

func (page *page) makeSeqReadFromFile(
	bufferedReader *bufio.Reader,
) interfaces.SeqError[*row] {
	return func(yield func(*row, error) bool) {
		if bufferedReader == nil {
			return
		}

		var row row

		for {
			n, err := page.writeIntoRow(&row, bufferedReader)
			if err != nil {
				if err == io.EOF {
					return
					// no-op
					// TODO why might this ever be the case
				} else if errors.Is(err, io.ErrUnexpectedEOF) && n == 0 {
					return
				}

				yield(nil, errors.Wrap(err))
				return
			}

			if !yield(&row, nil) {
				return
			}
		}
	}
}

func (page *page) flushNew(
	bufferedReader *bufio.Reader,
	bufferedWriter *bufio.Writer,
) (err error) {
	seq := quiter.MergeSeqLeft(
		quiter.MakeSeqErrorFromSeq(page.added.All()),
		page.makeSeqReadFromFile(bufferedReader),
		page.compareRowForFlush,
	)

	for row, errIter := range seq {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return err
		}

		if _, err = page.readFromRow(row, bufferedWriter); err != nil {
			err = errors.Wrap(errIter)
			return err
		}
	}

	return err
}
