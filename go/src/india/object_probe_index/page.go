package object_probe_index

import (
	"bufio"
	"io"
	"os"
	"sync"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/merkle_ids"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/golf/env_ui"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
)

type page struct {
	sync.Mutex     // for the buffered reader
	file           *os.File
	bufferedReader bufio.Reader
	added          *heap.Heap[row, *row]
	envRepo        env_repo.Env
	searchFunc     func(interfaces.BlobId) (mid int64, err error)
	id             page_id.PageId
}

func (page *page) initialize(
	equaler interfaces.Equaler[*row],
	envRepo env_repo.Env,
	pid page_id.PageId,
) (err error) {
	page.added = heap.Make(
		equaler,
		rowLessor{},
		rowResetter{},
	)

	page.envRepo = envRepo
	page.id = pid

	page.searchFunc = page.seekToFirstBinarySearch

	if err = page.open(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (page *page) open() (err error) {
	if page.file != nil {
		if err = page.file.Close(); err != nil {
			err = errors.Wrap(err)
			return
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

		return
	}

	return
}

func (page *page) GetObjectProbeIndexPage() pageInterface {
	return page
}

func (page *page) AddSha(sh interfaces.BlobId, loc Loc) (err error) {
	if sh.IsNull() {
		return
	}

	page.Lock()
	defer page.Unlock()

	return page.addSha(sh, loc)
}

func (page *page) addSha(sh interfaces.BlobId, loc Loc) (err error) {
	if sh.IsNull() {
		return
	}

	r := &row{
		Loc: loc,
	}

	if err = r.sha.SetDigest(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	page.added.Push(r)

	return
}

func (page *page) GetRowCount() (n int64, err error) {
	var fi os.FileInfo

	if fi, err = page.file.Stat(); err != nil {
		err = errors.Wrap(err)
		return
	}

	n = fi.Size()/RowSize - 1

	return
}

func (page *page) ReadOne(sh interfaces.BlobId) (loc Loc, err error) {
	page.Lock()
	defer page.Unlock()

	var start int64

	if start, err = page.searchFunc(sh); err != nil {
		if !collections.IsErrNotFound(err) {
			err = errors.Wrap(err)
		}

		return
	}

	if err = page.seekAndResetTo(start); err != nil {
		err = errors.Wrap(err)
		return
	}

	if loc, _, err = page.readCurrentLoc(sh, &page.bufferedReader); err != nil {
		err = errors.Wrapf(err, "Start: %d", start)
		return
	}

	return
}

func (page *page) ReadMany(sh interfaces.BlobId, locs *[]Loc) (err error) {
	page.Lock()
	defer page.Unlock()

	var start int64

	if start, err = page.searchFunc(sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = page.seekAndResetTo(start); err != nil {
		err = errors.Wrap(err)
		return
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
			return
		}

		if found {
			*locs = append(*locs, loc)
		}
	}

	return
}

func (page *page) readCurrentLoc(
	expectedBlobId interfaces.BlobId,
	bufferedReader *bufio.Reader,
) (out Loc, found bool, err error) {
	if expectedBlobId.IsNull() {
		err = errors.ErrorWithStackf("empty sha")
		return
	}

	if found, err = merkle_ids.EqualsReader(expectedBlobId, bufferedReader); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	} else if !found {
		err = io.EOF
		return
	}

	var n int64
	n, err = out.ReadFrom(bufferedReader)

	if n > 0 {
		found = true
	}

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (page *page) seekAndResetTo(loc int64) (err error) {
	if _, err = page.file.Seek(loc*RowSize, io.SeekStart); err != nil {
		err = errors.Wrap(err)
		return
	}

	page.bufferedReader.Reset(page.file)

	return
}

func (page *page) PrintAll(env env_ui.Env) (err error) {
	page.Lock()
	defer page.Unlock()

	if page.file == nil {
		return
	}

	if err = page.seekAndResetTo(0); err != nil {
		err = errors.Wrap(err)
		return
	}

	for {
		var current row

		if _, err = current.ReadFrom(&page.bufferedReader); err != nil {
			err = errors.WrapExceptSentinelAsNil(err, io.EOF)
			return
		}

		env.GetUI().Printf("%s", &current)
	}
}

func (page *page) Flush() (err error) {
	page.Lock()
	defer page.Unlock()

	if page.added.Len() == 0 {
		return
	}

	if page.file != nil {
		if err = page.seekAndResetTo(0); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var ft *os.File

	if ft, err = page.envRepo.GetTempLocal().FileTemp(); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, ft)

	w := bufio.NewWriter(ft)

	defer errors.DeferredFlusher(&err, w)

	var current row

	// TODO make iterator
	getOne := func() (r *row, err error) {
		if page.file == nil {
			err = io.EOF
			return
		}

		var n int64
		n, err = current.ReadFrom(&page.bufferedReader)
		if err != nil {
			if errors.IsEOF(err) {
				// no-op
				// TODO why might this ever be the case
			} else if errors.Is(err, io.ErrUnexpectedEOF) && n == 0 {
				err = io.EOF
			}

			err = errors.WrapExceptSentinel(err, io.EOF)
			return
		}

		r = &current

		return
	}

	if err = heap.MergeStream(
		page.added,
		func() (tz *row, err error) {
			tz, err = getOne()

			if errors.IsEOF(err) || tz == nil {
				err = errors.MakeErrStopIteration()
			}

			return
		},
		func(r *row) (err error) {
			_, err = r.WriteTo(w)
			return
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = os.Rename(
		ft.Name(),
		page.id.Path(),
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	page.added.Reset()

	if err = page.open(); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
