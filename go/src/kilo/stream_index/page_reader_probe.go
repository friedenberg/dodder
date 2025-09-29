package stream_index

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type probePageReader struct {
	pageId   page_id.PageId
	readerAt io.ReaderAt
	decoder  binaryDecoder
}

func (index *Index) makeProbePageReader(
	pageIndex PageIndex,
) (probePageReader, errors.FuncErr) {
	page := &index.pages[pageIndex]
	pageReader := probePageReader{
		pageId:  page.pageId,
		decoder: makeBinaryWithQueryGroup(nil, ids.SigilHistory),
	}

	var err error
	var blobReader interfaces.BlobReader

	if blobReader, err = index.envRepo.MakeNamedBlobReader(
		pageReader.pageId.Path(),
	); err != nil {
		if errors.IsNotExist(err) {
			return pageReader, func() error { return nil }
		} else {
			panic(err)
		}
	}

	pageReader.readerAt = blobReader

	return pageReader, func() (err error) {
		if err = blobReader.Close(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}
}

func (pageReader *probePageReader) readOneCursor(
	cursor ohio.Cursor,
	object *sku.Transacted,
) (err error) {
	// pages get deleted before reindexing, so this is actually valid to have a
	// non-nil cursor request
	if pageReader.readerAt == nil {
		err = collections.MakeErrNotFound(cursor)
		return err
	}

	var bytesRead int64

	objectPlus := objectWithCursorAndSigil{
		objectWithSigil: objectWithSigil{
			Transacted: object,
		},
		Cursor: cursor,
	}

	if bytesRead, err = pageReader.decoder.readFormatExactly(
		pageReader.readerAt,
		&objectPlus,
	); err != nil {
		if err == io.EOF {
			if bytesRead == cursor.ContentLength {
				err = nil
				goto NO_ERR
			} else {
				err = io.ErrUnexpectedEOF
			}
		}

		err = errors.Wrapf(
			err,
			"Range: %q, Page: %q, BytesRead: %d",
			cursor,
			pageReader.pageId.Path(),
			bytesRead,
		)

		return err
	}

NO_ERR:

	return err
}
