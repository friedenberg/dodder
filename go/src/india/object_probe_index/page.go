package object_probe_index

import (
	"bufio"
	"io"
	"os"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type page struct {
	sync.Mutex     // for the buffered reader
	hashType       markl.FormatHash
	rowWidth       int
	file           *os.File
	bufferedReader bufio.Reader
	added          *heap.Heap[row, *row]
	envRepo        env_repo.Env
	searchFunc     func(interfaces.MarklId) (mid int64, err error)
	id             page_id.PageId
}

func (page *page) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	pageId page_id.PageId,
	hashType markl.FormatHash,
	rowWidth int,
) (err error) {
	page.added = heap.Make(
		equaler,
		rowLessor{},
		rowResetter{},
	)

	page.envRepo = envRepo
	page.id = pageId
	page.hashType = hashType
	page.rowWidth = rowWidth

	page.searchFunc = page.seekToFirstBinarySearch

	if err = page.open(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (page *page) open() (err error) {
	if page.file != nil {
		if err = page.file.Close(); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if page.file, err = files.OpenFile(
		page.id.Path(),
		os.O_RDONLY,
		0o666,
	); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			err = errors.Wrap(err)
		}

		return err
	}

	return err
}

func (page *page) AddMarklId(id interfaces.MarklId, loc Loc) (err error) {
	if id.IsNull() {
		return err
	}

	page.Lock()
	defer page.Unlock()

	row := &row{
		Loc: loc,
	}

	if err = row.Digest.SetDigest(id); err != nil {
		err = errors.Wrap(err)
		return err
	}

	page.added.Push(row)

	return err
}

func (page *page) GetRowCount() (n int64, err error) {
	var fileInfo os.FileInfo

	if fileInfo, err = page.file.Stat(); err != nil {
		err = errors.Wrap(err)
		return n, err
	}

	n = fileInfo.Size()/int64(page.rowWidth) - 1

	return n, err
}

func (page *page) ReadOne(id interfaces.MarklId) (loc Loc, err error) {
	page.Lock()
	defer page.Unlock()

	var start int64

	if start, err = page.searchFunc(id); err != nil {
		if !collections.IsErrNotFound(err) {
			err = errors.Wrap(err)
		}

		return loc, err
	}

	if err = page.seekAndResetTo(start); err != nil {
		err = errors.Wrap(err)
		return loc, err
	}

	if loc, _, err = page.readCurrentLoc(id, &page.bufferedReader); err != nil {
		err = errors.Wrapf(err, "Start: %d", start)
		return loc, err
	}

	return loc, err
}

func (page *page) ReadMany(sh interfaces.MarklId, locs *[]Loc) (err error) {
	page.Lock()
	defer page.Unlock()

	var start int64

	if start, err = page.searchFunc(sh); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = page.seekAndResetTo(start); err != nil {
		err = errors.Wrap(err)
		return err
	}

	isEOF := false

	for !isEOF {
		var loc Loc
		var found bool

		loc, found, err = page.readCurrentLoc(sh, &page.bufferedReader)

		if err == io.EOF {
			err = nil
			isEOF = true
		} else if err != nil {
			err = errors.Wrap(err)
			return err
		}

		if found {
			*locs = append(*locs, loc)
		}
	}

	return err
}

func (page *page) readCurrentLoc(
	expectedBlobId interfaces.MarklId,
	bufferedReader *bufio.Reader,
) (out Loc, found bool, err error) {
	if expectedBlobId.IsNull() {
		err = errors.ErrorWithStackf("empty sha")
		return out, found, err
	}

	if found, err = markl.EqualsReader(expectedBlobId, bufferedReader); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return out, found, err
	} else if !found {
		err = io.EOF
		return out, found, err
	}

	var n int64
	n, err = out.ReadFrom(bufferedReader)

	if n > 0 {
		found = true
	}

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return out, found, err
	}

	return out, found, err
}

func (page *page) seekAndResetTo(loc int64) (err error) {
	if _, err = page.file.Seek(
		loc*int64(page.rowWidth),
		io.SeekStart,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	page.bufferedReader.Reset(page.file)

	return err
}

func (page *page) PrintAll(env env_ui.Env) (err error) {
	page.Lock()
	defer page.Unlock()

	if page.file == nil {
		return err
	}

	if err = page.seekAndResetTo(0); err != nil {
		err = errors.WrapExceptSentinelAsNil(err, io.EOF)
		return err
	}

	for {
		var current row

		if _, err = page.writeIntoRow(
			&current,
			&page.bufferedReader,
		); err != nil {
			err = errors.WrapExceptSentinelAsNil(err, io.EOF)
			return err
		}

		env.GetUI().Printf("%s", &current)
	}
}

func (page *page) Flush() (err error) {
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

	if err = page.flushOld(&page.bufferedReader, bufferedWriter); err != nil {
		err = errors.Wrap(err)
		return err
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

func (page *page) flushOld(
	bufferedReader *bufio.Reader,
	bufferedWriter *bufio.Writer,
) (err error) {
	var current row

	getOne := func() (row *row, err error) {
		if page.file == nil {
			err = io.EOF
			return row, err
		}

		var n int64
		n, err = page.writeIntoRow(&current, bufferedReader)
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

	if err = heap.MergeHeapAndRestore(
		page.added,
		func() (row *row, err error) {
			row, err = getOne()

			if errors.IsEOF(err) || row == nil {
				err = errors.MakeErrStopIteration()
			}

			return row, err
		},
		func(row *row) (err error) {
			_, err = page.readFromRow(row, bufferedWriter)
			return err
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
