package stream_index

import (
	"bufio"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/comments"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
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

	if pageReader.blobReader, err = env_repo.MakeNamedBlobReaderOrNullReader(
		pageReader.namedBlobAccess,
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

func makeSeqObjectWithCursorAndSigilFromReader(
	reader io.Reader,
	queryGroup sku.PrimitiveQueryGroup,
) interfaces.SeqError[objectWithCursorAndSigil] {
	return func(yield func(objectWithCursorAndSigil, error) bool) {
		decoder := makeBinaryWithQueryGroup(queryGroup, ids.SigilHistory)

		var object objectWithCursorAndSigil

		for {
			object.Offset += object.ContentLength
			object.Transacted = sku.GetTransactedPool().Get()

			var err error

			if object.ContentLength, err = decoder.readFormatAndMatchSigil(
				reader,
				&object,
			); err != nil {
				yield(object, errors.WrapExceptSentinelAsNil(err, io.EOF))
				return
			}

			if !yield(object, nil) {
				return
			}
		}
	}
}

func makeSeqObjectFromReader(
	reader io.Reader,
	queryGroup sku.PrimitiveQueryGroup,
) interfaces.SeqError[*sku.Transacted] {
	return func(yield func(*sku.Transacted, error) bool) {
		seq := makeSeqObjectWithCursorAndSigilFromReader(reader, queryGroup)
		for objectPlus, err := range seq {
			if err != nil {
				yield(nil, errors.Wrap(err))
				return
			}

			if !yield(objectPlus.Transacted, nil) {
				return
			}
		}
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
	includeAddedHistory bool
	includeAddedLatest  bool
}

func (pageReader *streamPageReader) makeSeq(
	query sku.PrimitiveQueryGroup,
	pageReadOptions pageReadOptions,
) interfaces.SeqError[*sku.Transacted] {
	if !pageReadOptions.includeAddedHistory &&
		!pageReadOptions.includeAddedLatest {
		return makeSeqObjectFromReader(
			pageReader.bufferedReader,
			query,
		)
	}

	return func(yield func(*sku.Transacted, error) bool) {
		comments.Optimize("determine performance of this")
		addedHistory := pageReader.additionsHistory.objects.Copy()

		{
			seq := quiter.MergeSequences(
				addedHistory.AllError(),
				makeSeqObjectFromReader(pageReader.bufferedReader, query),
				sku.TransactedCompare,
			)

			for object, errIter := range seq {
				if errIter != nil {
					yield(nil, errors.Wrap(errIter))
					return
				}

				if !yield(object, nil) {
					return
				}
			}
		}

		if !pageReadOptions.includeAddedLatest {
			return
		}

		comments.Optimize("determine performance of this")
		addedLatest := pageReader.additionsLatest.objects.Copy()

		{
			seq := quiter.MergeSequences(
				addedLatest.AllError(),
				quiter.MakeSeqErrorEmpty[*sku.Transacted](),
				sku.TransactedCompare,
			)

			for object, errIter := range seq {
				if errIter != nil {
					yield(nil, errors.Wrap(errIter))
					return
				}

				if !yield(object, nil) {
					return
				}
			}
		}

		return
	}
}
