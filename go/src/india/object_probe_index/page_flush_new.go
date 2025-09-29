package object_probe_index

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
)

func (page *page) makeSeqReadFromFile() interfaces.SeqError[*row] {
	return func(yield func(*row, error) bool) {
		if page.file == nil {
			return
		}

		var row row

		for {
			n, err := page.writeIntoRow(&row, &page.bufferedReader)
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

func (page *page) FlushNew() (err error) {
	page.Lock()
	defer page.Unlock()

	if page.added.Len() == 0 {
		return err
	}

	if page.file != nil {
		if err = page.seekAndResetTo(0); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	var temporaryFile *os.File

	if temporaryFile, err = page.envRepo.GetTempLocal().FileTemp(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, temporaryFile)

	bufferedWriter, repool := pool.GetBufferedWriter(temporaryFile)
	defer repool()

	defer errors.DeferredFlusher(&err, bufferedWriter)

	seq := quiter.MergeSequences(
		quiter.MakeSeqErrorFromSeq(page.added.All()),
		page.makeSeqReadFromFile(),
		page.compareRows,
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

	if err = os.Rename(
		temporaryFile.Name(),
		page.id.Path(),
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	page.added.Reset()

	if err = page.open(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
