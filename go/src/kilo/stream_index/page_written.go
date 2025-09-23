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
	page_id.PageId
	sunrise ids.Tai
	*probeIndex
	hasChanges          bool
	envRepo             env_repo.Env
	preWrite            interfaces.FuncIter[*sku.Transacted]
	config              store_config.Store
	addedObjectIdLookup map[string]struct{}

	added, addedLatest *sku.ListTransacted
}

func (page *writtenPage) initialize(
	pid page_id.PageId,
	index *Index,
) {
	page.envRepo = index.envRepo
	page.sunrise = index.sunrise
	page.PageId = pid
	page.added = sku.MakeListTransacted()
	page.addedLatest = sku.MakeListTransacted()
	page.preWrite = index.preWrite
	page.probeIndex = &index.probeIndex
	page.addedObjectIdLookup = make(map[string]struct{})
}

func (page *writtenPage) readOneRange(
	raynge object_probe_index.Cursor,
	object *sku.Transacted,
) (err error) {
	var file *os.File

	if file, err = files.Open(page.Path()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, file)

	bites := make([]byte, raynge.ContentLength)

	if _, err = file.ReadAt(bites, raynge.Offset); err != nil {
		err = errors.Wrapf(err, "Range: %q, Page: %q", raynge, page.PageId)
		return err
	}

	dec := makeBinaryWithQueryGroup(nil, ids.SigilHistory)

	skWR := objectWithCursorAndSigil{
		objectWithSigil: objectWithSigil{
			Transacted: object,
		},
		Cursor: raynge,
	}

	if _, err = dec.readFormatExactly(file, &skWR); err != nil {
		err = errors.Wrapf(
			err,
			"Range: %q, Page: %q",
			raynge,
			page.PageId.Path(),
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
	dec := makeBinaryWithQueryGroup(queryGroup, ids.SigilHistory)

	var sk objectWithCursorAndSigil

	for {
		sk.Offset += sk.ContentLength
		sk.Transacted = sku.GetTransactedPool().Get()
		sk.ContentLength, err = dec.readFormatAndMatchSigil(reader, &sk)
		if err != nil {
			if errors.IsEOF(err) {
				err = nil
			} else {
				err = errors.Wrap(err)
			}

			return err
		}

		if err = output(sk); err != nil {
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
	var reader io.ReadCloser

	if reader, err = page.envRepo.ReadCloserCache(page.Path()); err != nil {
		if errors.IsNotExist(err) {
			reader = io.NopCloser(bytes.NewReader(nil))
			err = nil
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	defer errors.DeferredCloser(&err, reader)

	bufferedReader, repool := pool.GetBufferedReader(reader)
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

	dec := makeBinaryWithQueryGroup(query, ids.SigilHistory)

	ui.TodoP3("determine performance of this")
	added := page.added.Copy()

	var object objectWithCursorAndSigil

	if err = heap.MergeStream(
		&added,
		func() (transacted *sku.Transacted, err error) {
			transacted = sku.GetTransactedPool().Get()
			object.Transacted = transacted

			_, err = dec.readFormatAndMatchSigil(bufferedReader, &object)
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
		func() (tz *sku.Transacted, err error) {
			err = errors.MakeErrStopIteration()
			return tz, err
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
) func() error {
	return func() (err error) {
		pageWriter := &pageWriter{
			pageId:      page.PageId,
			writtenPage: page,
			preWrite:    page.preWrite,
			envRepo:     page.envRepo,
			probeIndex:  page.probeIndex,
			path:        page.Path(),
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
