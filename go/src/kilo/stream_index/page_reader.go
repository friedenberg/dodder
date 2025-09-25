package stream_index

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type pageReader struct {
	*writtenPage
	blobReader     interfaces.BlobReader
	bufferedReader *bufio.Reader
	envRepo        env_repo.Env
}

func (index *Index) makePageReader(
	pageIndex PageIndex,
) (pageReader, errors.FuncErr) {
	pageReader := pageReader{
		writtenPage: &index.pages[pageIndex],
		envRepo:     index.envRepo,
	}

	var err error

	if pageReader.blobReader, err = pageReader.envRepo.MakeNamedBlobReader(
		pageReader.pageId.Path(),
	); err != nil {
		panic(err)
	}

	var repool interfaces.FuncRepool

	pageReader.bufferedReader, repool = pool.GetBufferedReader(
		pageReader.blobReader,
	)

	return pageReader, func() error {
		repool()
		return pageReader.blobReader.Close()
	}
}

func (pageReader *pageReader) readOneCursor(
	cursor object_probe_index.Cursor,
	object *sku.Transacted,
) (err error) {
	bites := make([]byte, cursor.ContentLength)

	if _, err = pageReader.blobReader.ReadAt(bites, cursor.Offset); err != nil {
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

	if _, err = decoder.readFormatExactly(
		pageReader.blobReader,
		&objectPlus,
	); err != nil {
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

func (pageReader *pageReader) readFrom(
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
			err = errors.WrapExceptSentinelAsNil(err, io.EOF)
			return err
		}

		if err = output(object); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}
}

type pageReadOptions struct {
	includeAdded       bool
	includeAddedLatest bool
}

func (pageReader *pageReader) readFull(
	query sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[*sku.Transacted],
	pageReadOptions pageReadOptions,
) (err error) {
	if !pageReadOptions.includeAdded && !pageReadOptions.includeAddedLatest {
		if err = pageReader.readFrom(
			pageReader.bufferedReader,
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
	added := pageReader.additions.added.Copy()

	var object objectWithCursorAndSigil

	if err = heap.MergeStream(
		&added,
		func() (transacted *sku.Transacted, err error) {
			transacted = sku.GetTransactedPool().Get()
			object.Transacted = transacted

			_, err = decoder.readFormatAndMatchSigil(
				pageReader.bufferedReader,
				&object,
			)
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

	if !pageReadOptions.includeAddedLatest {
		return err
	}

	addedLatest := pageReader.additions.addedLatest.Copy()

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
