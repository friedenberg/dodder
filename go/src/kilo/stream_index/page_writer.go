package stream_index

import (
	"bufio"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type ObjectIdToObject map[string]objectMetaWithCursorAndSigil

type pageWriter struct {
	envRepo     env_repo.Env
	pageId      page_id.PageId
	preWrite    interfaces.FuncIter[*sku.Transacted]
	writtenPage *writtenPage
	path        string

	binaryEncoder binaryEncoder

	file *os.File

	changesAreHistorical bool

	probeIndex *probeIndex

	cursor object_probe_index.Cursor

	latestObjects ObjectIdToObject
}

func (pageWriter *pageWriter) Flush() (err error) {
	if !pageWriter.writtenPage.hasChanges {
		ui.Log().Print("not flushing, no changes")
		return err
	}

	defer pageWriter.writtenPage.added.Reset()
	defer pageWriter.writtenPage.addedLatest.Reset()

	pageWriter.latestObjects = make(ObjectIdToObject)

	// If the cache file does not exist and we have nothing to add, short
	// circuit the flush. This condition occurs on the initial init when the
	// konfig is changed but there are no objects yet.
	if !files.Exists(pageWriter.path) &&
		pageWriter.writtenPage.waitingToAddLen() == 0 {
		return err
	}

	ui.Log().Print("changesAreHistorical", pageWriter.changesAreHistorical)
	ui.Log().Print("added", pageWriter.writtenPage.added.Len())
	ui.Log().Print("addedtail", pageWriter.writtenPage.addedLatest.Len())

	if pageWriter.writtenPage.added.Len() == 0 &&
		!pageWriter.changesAreHistorical {
		if pageWriter.file, err = files.OpenReadWrite(pageWriter.path); err != nil {
			err = errors.Wrap(err)
			return err
		}

		bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(
			pageWriter.file,
		)
		defer repoolBufferedWriter()

		defer errors.DeferredCloser(&err, pageWriter.file)

		bufferedReader, repoolBufferedReader := pool.GetBufferedReader(
			pageWriter.file,
		)
		defer repoolBufferedReader()

		return pageWriter.flushJustLatest(bufferedReader, bufferedWriter)
	} else {
		if pageWriter.file, err = pageWriter.envRepo.GetTempLocal().FileTemp(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		defer errors.DeferredCloseAndRename(
			&err,
			pageWriter.file,
			pageWriter.file.Name(),
			pageWriter.path,
		)

		bufferedWriter, repoolBufferedWriter := pool.GetBufferedWriter(pageWriter.file)
		defer repoolBufferedWriter()

		return pageWriter.flushBoth(bufferedWriter)
	}
}

func (pageWriter *pageWriter) flushBoth(
	bufferedWriter *bufio.Writer,
) (err error) {
	ui.Log().Printf("flushing both: %s", pageWriter.path)

	chain := quiter.MakeChain(
		pageWriter.preWrite,
		pageWriter.makeWriteOne(bufferedWriter),
	)

	if err = pageWriter.writtenPage.copyJustHistoryAndAdded(
		sku.MakePrimitiveQueryGroup(),
		chain,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for {
		popped, ok := pageWriter.writtenPage.addedLatest.Pop()

		if !ok {
			break
		}

		if err = chain(popped); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, object := range pageWriter.latestObjects {
		if err = pageWriter.updateSigilWithLatest(object); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (pageWriter *pageWriter) updateSigilWithLatest(
	objectMeta objectMetaWithCursorAndSigil,
) (err error) {
	objectMeta.Add(ids.SigilLatest)

	if err = pageWriter.binaryEncoder.updateSigil(
		pageWriter.file,
		objectMeta.Sigil,
		objectMeta.Offset,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (pageWriter *pageWriter) flushJustLatest(
	bufferedReader *bufio.Reader,
	bufferedWriter *bufio.Writer,
) (err error) {
	ui.Log().Printf("flushing just tail: %s", pageWriter.path)

	if err = pageWriter.writtenPage.copyJustHistoryFrom(
		bufferedReader,
		sku.MakePrimitiveQueryGroup(),
		func(object objectWithCursorAndSigil) (err error) {
			pageWriter.cursor = object.Cursor
			pageWriter.saveToLatestMap(object.Transacted, object.Sigil)
			return err
		},
	); err != nil {
		err = errors.Wrapf(err, "Page: %s", pageWriter.pageId)
		return err
	}

	chain := quiter.MakeChain(
		pageWriter.preWrite,
		pageWriter.removeOldLatest,
		pageWriter.makeWriteOne(bufferedWriter),
	)

	for {
		popped, ok := pageWriter.writtenPage.addedLatest.Pop()

		if !ok {
			break
		}

		if err = chain(popped); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	if err = bufferedWriter.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, object := range pageWriter.latestObjects {
		if err = pageWriter.updateSigilWithLatest(object); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (pageWriter *pageWriter) makeWriteOne(
	bufferedWriter *bufio.Writer,
) interfaces.FuncIter[*sku.Transacted] {
	return func(object *sku.Transacted) (err error) {
		// defer func() {
		// 	r := recover()

		// 	if r == nil {
		// 		return
		// 	}

		// 	ui.Debug().Print(z)
		// 	panic(r)
		// }()
		pageWriter.cursor.Offset += pageWriter.cursor.ContentLength

		objectOld := pageWriter.latestObjects[object.GetObjectId().String()]

		object.Metadata.Cache.ParentTai = objectOld.Tai

		if pageWriter.cursor.ContentLength, err = pageWriter.binaryEncoder.writeFormat(
			bufferedWriter,
			objectWithSigil{Transacted: object},
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = pageWriter.saveToLatestMap(object, ids.SigilHistory); err != nil {
			err = errors.Wrap(err)
			return err
		}

		if err = pageWriter.probeIndex.saveOneObjectLoc(
			object,
			object_probe_index.Loc{
				Page:   pageWriter.pageId.Index,
				Cursor: pageWriter.cursor,
			},
		); err != nil {
			err = errors.Wrap(err)
			return err
		}

		return err
	}
}

func (pageWriter *pageWriter) saveToLatestMap(
	object *sku.Transacted,
	sigil ids.Sigil,
) (err error) {
	objectId := object.GetObjectId()
	objectIdString := objectId.String()

	objectOld := pageWriter.latestObjects[objectIdString]
	objectOld.Cursor = pageWriter.cursor
	objectOld.Tai = object.GetTai()
	objectOld.Sigil = sigil

	if object.Metadata.Cache.Dormant.Bool() {
		objectOld.Add(ids.SigilHidden)
	} else {
		objectOld.Del(ids.SigilHidden)
	}

	pageWriter.latestObjects[objectIdString] = objectOld

	return err
}

func (pageWriter *pageWriter) removeOldLatest(
	objectLatest *sku.Transacted,
) (err error) {
	objectIdString := objectLatest.ObjectId.String()
	objectOld, ok := pageWriter.latestObjects[objectIdString]

	if !ok {
		return err
	}

	objectOld.Del(ids.SigilLatest)

	if err = pageWriter.binaryEncoder.updateSigil(
		pageWriter.file,
		objectOld.Sigil,
		objectOld.Offset,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
