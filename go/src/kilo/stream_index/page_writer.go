package stream_index

import (
	"bufio"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/quiter"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
)

type ObjectIdToObject map[string]skuWithRangeAndSigil

type pageWriter struct {
	writtenPage *writtenPage

	binaryDecoder
	binaryEncoder

	*os.File

	bufio.Reader
	bufio.Writer

	changesAreHistorical bool

	*probeIndex

	object_probe_index.Range

	ObjectIdToObjectMap ObjectIdToObject
}

func (pageWriter *pageWriter) Flush() (err error) {
	if !pageWriter.writtenPage.hasChanges {
		ui.Log().Print("not flushing, no changes")
		return err
	}

	defer pageWriter.writtenPage.added.Reset()
	defer pageWriter.writtenPage.addedLatest.Reset()

	pageWriter.ObjectIdToObjectMap = make(ObjectIdToObject)
	pageWriter.binaryDecoder = makeBinary(ids.SigilHistory)
	pageWriter.binaryDecoder.sigil = ids.SigilHistory

	path := pageWriter.writtenPage.Path()

	// If the cache file does not exist and we have nothing to add, short
	// circuit the flush. This condition occurs on the initial init when the
	// konfig is changed but there are no zettels yet.
	if !files.Exists(path) && pageWriter.writtenPage.waitingToAddLen() == 0 {
		return err
	}

	ui.Log().Print("changesAreHistorical", pageWriter.changesAreHistorical)
	ui.Log().Print("added", pageWriter.writtenPage.added.Len())
	ui.Log().Print("addedtail", pageWriter.writtenPage.addedLatest.Len())

	if pageWriter.writtenPage.added.Len() == 0 &&
		!pageWriter.changesAreHistorical {
		if pageWriter.File, err = files.OpenReadWrite(path); err != nil {
			err = errors.Wrap(err)
			return err
		}

		defer errors.DeferredCloser(&err, pageWriter.File)

		pageWriter.Reader.Reset(pageWriter.File)
		pageWriter.Writer.Reset(pageWriter.File)

		return pageWriter.flushJustLatest()
	} else {
		if pageWriter.File, err = pageWriter.writtenPage.envRepo.GetTempLocal().FileTemp(); err != nil {
			err = errors.Wrap(err)
			return err
		}

		defer errors.DeferredCloseAndRename(&err, pageWriter.File, pageWriter.Name(), path)

		pageWriter.Reader.Reset(pageWriter.File)
		pageWriter.Writer.Reset(pageWriter.File)

		return pageWriter.flushBoth()
	}
}

func (pageWriter *pageWriter) flushBoth() (err error) {
	ui.Log().Printf("flushing both: %s", pageWriter.writtenPage.Path())

	chain := quiter.MakeChain(
		pageWriter.writtenPage.preWrite,
		pageWriter.writeOne,
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

	if err = pageWriter.Writer.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, st := range pageWriter.ObjectIdToObjectMap {
		if err = pageWriter.updateSigilWithLatest(st); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (pageWriter *pageWriter) updateSigilWithLatest(
	st skuWithRangeAndSigil,
) (err error) {
	st.Add(ids.SigilLatest)

	if err = pageWriter.updateSigil(pageWriter, st.Sigil, st.Offset); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (pageWriter *pageWriter) flushJustLatest() (err error) {
	ui.Log().Printf("flushing just tail: %s", pageWriter.writtenPage.Path())

	if err = pageWriter.writtenPage.copyJustHistoryFrom(
		&pageWriter.Reader,
		sku.MakePrimitiveQueryGroup(),
		func(sk skuWithRangeAndSigil) (err error) {
			pageWriter.Range = sk.Range
			pageWriter.saveToLatestMap(sk.Transacted, sk.Sigil)
			return err
		},
	); err != nil {
		err = errors.Wrapf(err, "Page: %s", pageWriter.writtenPage.PageId)
		return err
	}

	chain := quiter.MakeChain(
		pageWriter.writtenPage.preWrite,
		pageWriter.removeOldLatest,
		pageWriter.writeOne,
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

	if err = pageWriter.Writer.Flush(); err != nil {
		err = errors.Wrap(err)
		return err
	}

	for _, st := range pageWriter.ObjectIdToObjectMap {
		if err = pageWriter.updateSigilWithLatest(st); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

func (pageWriter *pageWriter) writeOne(
	object *sku.Transacted,
) (err error) {
	// defer func() {
	// 	r := recover()

	// 	if r == nil {
	// 		return
	// 	}

	// 	ui.Debug().Print(z)
	// 	panic(r)
	// }()
	pageWriter.Offset += pageWriter.ContentLength

	previous := pageWriter.ObjectIdToObjectMap[object.GetObjectId().String()]

	if previous.Transacted != nil {
		object.Metadata.Cache.ParentTai = previous.GetTai()
	}

	if pageWriter.ContentLength, err = pageWriter.writeFormat(
		&pageWriter.Writer,
		skuWithSigil{Transacted: object},
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
			Page:  pageWriter.writtenPage.PageId.Index,
			Range: pageWriter.Range,
		},
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (pageWriter *pageWriter) saveToLatestMap(
	z *sku.Transacted,
	sigil ids.Sigil,
) (err error) {
	k := z.GetObjectId()
	ks := k.String()

	record := pageWriter.ObjectIdToObjectMap[ks]
	record.Range = pageWriter.Range

	if record.Transacted == nil {
		record.Transacted = sku.GetTransactedPool().Get()
	}

	sku.TransactedResetter.ResetWith(record.Transacted, z)

	record.Sigil = sigil

	if z.Metadata.Cache.Dormant.Bool() {
		record.Add(ids.SigilHidden)
	} else {
		record.Del(ids.SigilHidden)
	}

	pageWriter.ObjectIdToObjectMap[ks] = record

	return err
}

func (pageWriter *pageWriter) removeOldLatest(sk *sku.Transacted) (err error) {
	ks := sk.ObjectId.String()
	st, ok := pageWriter.ObjectIdToObjectMap[ks]

	if !ok {
		return err
	}

	st.Del(ids.SigilLatest)

	if err = pageWriter.updateSigil(pageWriter, st.Sigil, st.Offset); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
