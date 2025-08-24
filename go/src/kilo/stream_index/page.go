package stream_index

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/page_id"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

type Page struct {
	page_id.PageId
	sunrise ids.Tai
	*probe_index
	added, addedLatest *sku.ListTransacted
	hasChanges         bool
	envRepo            env_repo.Env
	preWrite           interfaces.FuncIter[*sku.Transacted]
	config             store_config.Store
	oids               map[string]struct{}
}

func (page *Page) initialize(
	pid page_id.PageId,
	i *Index,
) {
	page.envRepo = i.envRepo
	page.sunrise = i.sunrise
	page.PageId = pid
	page.added = sku.MakeListTransacted()
	page.addedLatest = sku.MakeListTransacted()
	page.preWrite = i.preWrite
	page.probe_index = &i.probe_index
	page.oids = make(map[string]struct{})
}

func (page *Page) readOneRange(
	raynge object_probe_index.Range,
	object *sku.Transacted,
) (err error) {
	var file *os.File

	if file, err = files.Open(page.Path()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, file)

	bites := make([]byte, raynge.ContentLength)

	if _, err = file.ReadAt(bites, raynge.Offset); err != nil {
		err = errors.Wrapf(err, "Range: %q, Page: %q", raynge, page.PageId)
		return
	}

	dec := makeBinaryWithQueryGroup(nil, ids.SigilHistory)

	skWR := skuWithRangeAndSigil{
		skuWithSigil: skuWithSigil{
			Transacted: object,
		},
		Range: raynge,
	}

	if _, err = dec.readFormatExactly(file, &skWR); err != nil {
		err = errors.Wrapf(
			err,
			"Range: %q, Page: %q",
			raynge,
			page.PageId.Path(),
		)
		return
	}

	return
}

func (page *Page) add(
	object *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	page.oids[object.ObjectId.String()] = struct{}{}
	objectClone := object.CloneTransacted()

	if page.sunrise.Less(objectClone.GetTai()) ||
		options.StreamIndexOptions.ForceLatest {
		page.addedLatest.Add(objectClone)
	} else {
		page.added.Add(objectClone)
	}

	page.hasChanges = true

	return
}

func (page *Page) waitingToAddLen() int {
	return page.added.Len() + page.addedLatest.Len()
}

func (page *Page) copyJustHistoryFrom(
	reader io.Reader,
	queryGroup sku.PrimitiveQueryGroup,
	output interfaces.FuncIter[skuWithRangeAndSigil],
) (err error) {
	dec := makeBinaryWithQueryGroup(queryGroup, ids.SigilHistory)

	var sk skuWithRangeAndSigil

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

			return
		}

		if err = output(sk); err != nil {
			err = errors.Wrap(err)
			return
		}
	}
}

func (page *Page) copyJustHistoryAndAdded(
	s sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return page.copyHistoryAndMaybeLatest(s, w, true, false)
}

func (page *Page) copyHistoryAndMaybeLatest(
	qg sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
	includeAdded bool,
	includeAddedLatest bool,
) (err error) {
	var r io.ReadCloser

	if r, err = page.envRepo.ReadCloserCache(page.Path()); err != nil {
		if errors.IsNotExist(err) {
			r = io.NopCloser(bytes.NewReader(nil))
			err = nil
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	defer errors.DeferredCloser(&err, r)

	br := bufio.NewReader(r)

	if !includeAdded && !includeAddedLatest {
		if err = page.copyJustHistoryFrom(
			br,
			qg,
			func(sk skuWithRangeAndSigil) (err error) {
				if err = w(sk.Transacted); err != nil {
					err = errors.Wrapf(err, "%s", sk.Transacted)
					return
				}

				return
			},
		); err != nil {
			err = errors.Wrap(err)
			return
		}

		return
	}

	dec := makeBinaryWithQueryGroup(qg, ids.SigilHistory)

	ui.TodoP3("determine performance of this")
	added := page.added.Copy()

	var sk skuWithRangeAndSigil

	if err = heap.MergeStream(
		&added,
		func() (tz *sku.Transacted, err error) {
			tz = sku.GetTransactedPool().Get()
			sk.Transacted = tz
			_, err = dec.readFormatAndMatchSigil(br, &sk)
			if err != nil {
				if errors.IsEOF(err) {
					err = errors.MakeErrStopIteration()
				} else {
					err = errors.Wrap(err)
				}

				return
			}

			return
		},
		w,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if !includeAddedLatest {
		return
	}

	addedLatest := page.addedLatest.Copy()

	if err = heap.MergeStream(
		&addedLatest,
		func() (tz *sku.Transacted, err error) {
			err = errors.MakeErrStopIteration()
			return
		},
		w,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (page *Page) MakeFlush(
	changesAreHistorical bool,
) func() error {
	return func() (err error) {
		pw := &writer{
			Page:        page,
			probe_index: page.probe_index,
		}

		if changesAreHistorical {
			pw.changesAreHistorical = true
			pw.hasChanges = true
		}

		if err = pw.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}

		page.hasChanges = false

		return
	}
}
