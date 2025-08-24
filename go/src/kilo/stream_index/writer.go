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

type ObjectIdShaMap map[string]skuWithRangeAndSigil

type writer struct {
	*Page
	binaryDecoder
	binaryEncoder
	*os.File
	bufio.Reader
	bufio.Writer

	changesAreHistorical bool

	*probe_index
	object_probe_index.Range
	offsetLast     int64
	ObjectIdShaMap ObjectIdShaMap
}

func (pw *writer) Flush() (err error) {
	if !pw.hasChanges {
		ui.Log().Print("not flushing, no changes")
		return
	}

	defer pw.added.Reset()
	defer pw.addedLatest.Reset()

	pw.ObjectIdShaMap = make(ObjectIdShaMap)
	pw.binaryDecoder = makeBinary(ids.SigilHistory)
	pw.binaryDecoder.sigil = ids.SigilHistory

	path := pw.Path()

	// If the cache file does not exist and we have nothing to add, short
	// circuit the flush. This condition occurs on the initial init when the
	// konfig is changed but there are no zettels yet.
	if !files.Exists(path) && pw.waitingToAddLen() == 0 {
		return
	}

	ui.Log().Print("changesAreHistorical", pw.changesAreHistorical)
	ui.Log().Print("added", pw.added.Len())
	ui.Log().Print("addedtail", pw.addedLatest.Len())

	if pw.added.Len() == 0 && !pw.changesAreHistorical {
		if pw.File, err = files.OpenReadWrite(path); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloser(&err, pw.File)

		pw.Reader.Reset(pw.File)
		pw.Writer.Reset(pw.File)

		return pw.flushJustLatest()
	} else {
		if pw.File, err = pw.Page.envRepo.GetTempLocal().FileTemp(); err != nil {
			err = errors.Wrap(err)
			return
		}

		defer errors.DeferredCloseAndRename(&err, pw.File, pw.Name(), path)

		pw.Reader.Reset(pw.File)
		pw.Writer.Reset(pw.File)

		return pw.flushBoth()
	}
}

func (pw *writer) flushBoth() (err error) {
	ui.Log().Printf("flushing both: %s", pw.Path())

	chain := quiter.MakeChain(
		pw.preWrite,
		pw.writeOne,
	)

	if err = pw.copyJustHistoryAndAdded(
		sku.MakePrimitiveQueryGroup(),
		chain,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	for {
		popped, ok := pw.addedLatest.Pop()

		if !ok {
			break
		}

		if err = chain(popped); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = pw.Writer.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, st := range pw.ObjectIdShaMap {
		if err = pw.updateSigilWithLatest(st); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (pw *writer) updateSigilWithLatest(st skuWithRangeAndSigil) (err error) {
	st.Add(ids.SigilLatest)

	if err = pw.updateSigil(pw, st.Sigil, st.Offset); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (pw *writer) flushJustLatest() (err error) {
	ui.Log().Printf("flushing just tail: %s", pw.Path())

	if err = pw.copyJustHistoryFrom(
		&pw.Reader,
		sku.MakePrimitiveQueryGroup(),
		func(sk skuWithRangeAndSigil) (err error) {
			pw.Range = sk.Range
			pw.saveToLatestMap(sk.Transacted, sk.Sigil)
			return
		},
	); err != nil {
		err = errors.Wrapf(err, "Page: %s", pw.PageId)
		return
	}

	chain := quiter.MakeChain(
		pw.preWrite,
		pw.removeOldLatest,
		pw.writeOne,
	)

	for {
		popped, ok := pw.addedLatest.Pop()

		if !ok {
			break
		}

		if err = chain(popped); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	if err = pw.Writer.Flush(); err != nil {
		err = errors.Wrap(err)
		return
	}

	for _, st := range pw.ObjectIdShaMap {
		if err = pw.updateSigilWithLatest(st); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (pw *writer) writeOne(
	z *sku.Transacted,
) (err error) {
	// defer func() {
	// 	r := recover()

	// 	if r == nil {
	// 		return
	// 	}

	// 	ui.Debug().Print(z)
	// 	panic(r)
	// }()
	pw.Offset += pw.ContentLength

	previous := pw.ObjectIdShaMap[z.GetObjectId().String()]

	if previous.Transacted != nil {
		z.Metadata.Cache.ParentTai = previous.GetTai()
	}

	if pw.ContentLength, err = pw.writeFormat(
		&pw.Writer,
		skuWithSigil{Transacted: z},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = pw.saveToLatestMap(z, ids.SigilHistory); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = pw.probe_index.saveOneLoc(
		z,
		object_probe_index.Loc{
			Page:  pw.PageId.Index,
			Range: pw.Range,
		},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (pw *writer) saveToLatestMap(
	z *sku.Transacted,
	sigil ids.Sigil,
) (err error) {
	k := z.GetObjectId()
	ks := k.String()

	record := pw.ObjectIdShaMap[ks]
	record.Range = pw.Range

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

	pw.ObjectIdShaMap[ks] = record

	return
}

func (pw *writer) removeOldLatest(sk *sku.Transacted) (err error) {
	ks := sk.ObjectId.String()
	st, ok := pw.ObjectIdShaMap[ks]

	if !ok {
		return
	}

	st.Del(ids.SigilLatest)

	if err = pw.updateSigil(pw, st.Sigil, st.Offset); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
