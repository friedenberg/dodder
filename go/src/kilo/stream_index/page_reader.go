package stream_index

import (
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type pageReader struct {
	*writtenPage
	envRepo env_repo.Env
}

func (index *Index) makePageReader(pageIndex PageIndex) pageReader {
	return pageReader{
		writtenPage: &index.pages[pageIndex],
		envRepo:     index.envRepo,
	}
}

func (pageReader *pageReader) readOneCursor(
	cursor object_probe_index.Cursor,
	object *sku.Transacted,
) (err error) {
	var file *os.File

	if file, err = files.Open(pageReader.pageId.Path()); err != nil {
		err = errors.Wrap(err)
		return err
	}

	defer errors.DeferredCloser(&err, file)

	bites := make([]byte, cursor.ContentLength)

	if _, err = file.ReadAt(bites, cursor.Offset); err != nil {
		err = errors.Wrapf(
			err,
			"Range: %q, Page: %q",
			cursor,
			pageReader.pageId,
		)
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
			pageReader.pageId.Path(),
		)
		return err
	}

	return err
}

func (pageReader *pageReader) copyJustHistoryFrom(
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

func (pageReader *pageReader) copyJustHistoryAndAdded(
	query sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return pageReader.copyHistoryAndMaybeLatest(query, output, true, false)
}

func (pageReader *pageReader) copyHistoryAndMaybeLatest(
	query sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
	includeAdded bool,
	includeAddedLatest bool,
) (err error) {
	var namedBlobReader io.ReadCloser

	if namedBlobReader, err = pageReader.envRepo.MakeNamedBlobReader(
		pageReader.pageId.Path(),
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
		if err = pageReader.copyJustHistoryFrom(
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
	added := pageReader.added.Copy()

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

	addedLatest := pageReader.addedLatest.Copy()

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
