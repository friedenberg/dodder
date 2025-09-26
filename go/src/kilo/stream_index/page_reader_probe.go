package stream_index

import (
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type probePageReader struct {
	*page
	// TODO switch to blob reader once debugged
	// blobReader     interfaces.BlobReader
	file    *os.File
	envRepo env_repo.Env
}

func (index *Index) makeProbePageReader(
	pageIndex PageIndex,
) (probePageReader, errors.FuncErr) {
	pageReader := probePageReader{
		page:    &index.pages[pageIndex],
		envRepo: index.envRepo,
	}

	var err error

	// if pageReader.blobReader, err = pageReader.envRepo.MakeNamedBlobReader(
	// 	pageReader.pageId.Path(),
	// ); err != nil {
	// 	panic(err)
	// }

	if pageReader.file, err = files.Open(pageReader.pageId.Path()); err != nil {
		if errors.IsNotExist(err) {
			err = nil
		} else {
			panic(err)
		}
	}

	return pageReader, func() (err error) {
		if pageReader.file != nil {
			if err = pageReader.file.Close(); err != nil {
				err = errors.Wrap(err)
				return err
			}
		}

		return err
		// return pageReader.blobReader.Close()
	}
}

func (pageReader *probePageReader) readOneCursor(
	cursor object_probe_index.Cursor,
	object *sku.Transacted,
) (err error) {
	// pages get deleted before reindexing, so this is actually valid to have a
	// non-nil cursor request
	if pageReader.file == nil {
		err = collections.MakeErrNotFound(cursor)
		return err
	}

	var bytesRead int64

	decoder := makeBinaryWithQueryGroup(nil, ids.SigilHistory)

	objectPlus := objectWithCursorAndSigil{
		objectWithSigil: objectWithSigil{
			Transacted: object,
		},
		Cursor: cursor,
	}

	if bytesRead, err = decoder.readFormatExactly(
		pageReader.file,
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
