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
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type streamPageReader struct {
	*page
	blobReader      interfaces.BlobReader
	bufferedReader  *bufio.Reader
	namedBlobAccess interfaces.NamedBlobAccess
}

func (index *Index) makeStreamPageReader(
	pageIndex PageIndex,
) (streamPageReader, errors.FuncErr) {
	pageReader := streamPageReader{
		page:            &index.pages[pageIndex],
		namedBlobAccess: index.envRepo,
	}

	var err error

	if pageReader.blobReader, err = pageReader.namedBlobAccess.MakeNamedBlobReader(
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

func (pageReader *streamPageReader) readFrom(
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

func (pageReader *streamPageReader) readFull(
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
