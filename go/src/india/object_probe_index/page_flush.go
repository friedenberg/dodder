package object_probe_index

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/cmp"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
)

func (page *page) makePullRowFromFile(
	bufferedReader *bufio.Reader,
) func() (*row, error) {
	var current row

	return func() (row *row, err error) {
		if bufferedReader == nil {
			err = io.EOF
			return row, err
		}

		var n int64
		n, err = page.readRowFrom(&current, bufferedReader)
		if err != nil {
			if errors.IsEOF(err) {
				// no-op
				// TODO why might this ever be the case
			} else if errors.Is(err, io.ErrUnexpectedEOF) && n == 0 {
				err = io.EOF
			}

			err = errors.WrapExceptSentinel(err, io.EOF)
			return row, err
		}

		row = &current

		return row, err
	}
}

func (page *page) flushOld(
	bufferedReader *bufio.Reader,
	bufferedWriter *bufio.Writer,
) (err error) {
	getOne := page.makePullRowFromFile(bufferedReader)

	if err = heap.MergeHeapAndRestore(
		page.added,
		func() (row *row, err error) {
			row, err = getOne()

			lastRow := errors.IsEOF(err) || row == nil

			if lastRow {
				err = errors.MakeErrStopIteration()
			}

			return row, err
		},
		func(row *row) (err error) {
			_, err = page.writeRowTo(row, bufferedWriter)
			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// confirmed this worked correctly
func (page *page) makeSeqReadFromFile(
	bufferedReader *bufio.Reader,
) interfaces.SeqError[*row] {
	return func(yield func(*row, error) bool) {
		if bufferedReader == nil {
			return
		}

		var row row

		for {
			n, err := page.readRowFrom(&row, bufferedReader)

			if err == io.EOF && n == 0 {
				return
			} else if err == io.EOF && n > 0 {
				yield(nil, errors.Wrap(io.ErrUnexpectedEOF))
				return
			} else if err != nil {
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
		func(left, right *row) cmp.Result {
			return cmp.Bytes(left.Digest.GetBytes(), right.Digest.GetBytes())
		},
	)

	for row, errIter := range seq {
		if errIter != nil {
			err = errors.Wrap(errIter)
			return err
		}

		if _, err = page.writeRowTo(row, bufferedWriter); err != nil {
			err = errors.WrapExceptSentinel(errIter, io.EOF)
			return err
		}
	}

	return err
}
