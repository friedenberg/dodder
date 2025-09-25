package stream_index

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

type writtenPage struct {
	pageId              page_id.PageId
	sunrise             ids.Tai
	probeIndex          *probeIndex
	hasChanges          bool
	envRepo             env_repo.Env
	config              store_config.Store
	addedObjectIdLookup map[string]struct{}

	added, addedLatest *sku.ListTransacted
}

func (page *writtenPage) initialize(
	pageId page_id.PageId,
	index *Index,
) {
	page.envRepo = index.envRepo
	page.sunrise = index.sunrise
	page.pageId = pageId
	page.added = sku.MakeListTransacted()
	page.addedLatest = sku.MakeListTransacted()
	page.probeIndex = &index.probeIndex
	page.addedObjectIdLookup = make(map[string]struct{})
}

func (page *writtenPage) readOneCursor(
	cursor object_probe_index.Cursor,
	object *sku.Transacted,
) (err error) {
	var file *os.File

	if file, err = files.Open(page.pageId.Path()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, file)

	bites := make([]byte, cursor.ContentLength)

	if _, err = file.ReadAt(bites, cursor.Offset); err != nil {
		err = errors.Wrapf(err, "Range: %q, Page: %q", cursor, page.pageId)
		return err
	}

	decoder := makeBinaryWithQueryGroup(nil, ids.SigilHistory)

	objectPlus := objectWithCursorAndSigil{
		objectWithSigil: objectWithSigil{
			Transacted: object,
		},
		Cursor: cursor,
	}

	if _, err = decoder.readFormatExactly(file, &objectPlus); err != nil {
		err = errors.Wrapf(
			err,
			"Range: %q, Page: %q",
			cursor,
			page.pageId.Path(),
		)
		return err
	}

	return err
}

// TODO write binary representation to file-backed buffered writer and then
// merge streams using raw binary data
func (page *writtenPage) add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	page.addedObjectIdLookup[object.ObjectId.String()] = struct{}{}
	objectClone := object.CloneTransacted()

	if page.sunrise.Less(objectClone.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		page.addedLatest.Add(objectClone)
	} else {
		page.added.Add(objectClone)
	}

	page.hasChanges = true

	return err
}

func (page *writtenPage) waitingToAddLen() int {
	return page.added.Len() + page.addedLatest.Len()
}

func (page *writtenPage) copyJustHistoryFrom(
	reader io.Reader,
	queryGroup sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[objectWithCursorAndSigil],
) (err error) {
	decoder := makeBinaryWithQueryGroup(queryGroup, ids.SigilHistory)

	var object objectWithCursorAndSigil

	for {
		object.Offset += object.ContentLength
		object.Transacted = sku.GetTransactedPool().Get()
		object.ContentLength, err = decoder.readFormatAndMatchSigil(
			reader,
			&object,
		)
		if err != nil {
			if errors.IsEOF(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}

		if err = output(object); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}
}

func (page *writtenPage) copyJustHistoryAndAdded(
	query sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return page.copyHistoryAndMaybeLatest(query, output, true, false)
}

func (page *writtenPage) copyHistoryAndMaybeLatest(
	query sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
	includeAdded bool,
	includeAddedLatest bool,
) (err error) {
	var namedBlobReader io.ReadCloser

	if namedBlobReader, err = page.envRepo.MakeNamedBlobReader(
		page.pageId.Path(),
	); err != nil {
		if errors.IsNotExist(err) {
			namedBlobReader = io.NopCloser(bytes.NewReader(nil))
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	defer errors.DeferredCloser(&err, namedBlobReader)

	bufferedReader, repool := pool.GetBufferedReader(namedBlobReader)
	defer repool()

	if !includeAdded && !includeAddedLatest {
		if err = page.copyJustHistoryFrom(
			bufferedReader,
			query,
			func(object objectWithCursorAndSigil) (err error) {
				if err = output(object.Transacted); err != nil {
					err = errors.Wrapf(err, "%s", object.Transacted)
					return err
				}

				return err
			},
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}

	decoder := makeBinaryWithQueryGroup(query, ids.SigilHistory)

	ui.TodoP3("determine performance of this")
	added := page.added.Copy()

	var object objectWithCursorAndSigil

	if err = heap.MergeStream(
		&added,
		func() (transacted *sku.Transacted, err error) {
			transacted = sku.GetTransactedPool().Get()
			object.Transacted = transacted

			_, err = decoder.readFormatAndMatchSigil(bufferedReader, &object)
			if err != nil {
				if errors.IsEOF(err) {
					err = errors.MakeErrStopIteration()
				} else {
					err = errors.Wrap(err)
				}

				return transacted, err
			}

			return transacted, err
		},
		output,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if !includeAddedLatest {
		return err
	}

	addedLatest := page.addedLatest.Copy()

	if err = heap.MergeStream(
		&addedLatest,
		func() (object *sku.Transacted, err error) {
			err = errors.MakeErrStopIteration()
			return object, err
		},
		output,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (page *writtenPage) MakeFlush(
	changesAreHistorical bool,
	preWrite interfaces.FuncIter[*sku.Transacted],
) func() error {
	return func() (err error) {
		pageWriter := &pageWriter{
			pageId:      page.pageId,
			writtenPage: page,
			preWrite:    preWrite,
			envRepo:     page.envRepo,
			probeIndex:  page.probeIndex,
			path:        page.pageId.Path(),
		}

		if changesAreHistorical {
			pageWriter.changesAreHistorical = true
			pageWriter.writtenPage.hasChanges = true
		}

		if err = pageWriter.Flush(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		page.hasChanges = false

		return err
	}
}
