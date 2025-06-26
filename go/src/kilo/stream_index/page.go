package stream_index

import (
	"bufio"
	"bytes"
	"io"
	"os"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
	"code.linenisgreat.com/dodder/go/src/delta/heap"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
	"code.linenisgreat.com/dodder/go/src/echo/ids"
	"code.linenisgreat.com/dodder/go/src/hotel/env_repo"
	"code.linenisgreat.com/dodder/go/src/india/object_probe_index"
	"code.linenisgreat.com/dodder/go/src/juliett/sku"
	"code.linenisgreat.com/dodder/go/src/mike/store_config"
)

type PageId = sha.PageId

type Page struct {
	PageId
	sunrise ids.Tai
	*probe_index
	added, addedLatest *sku.List
	hasChanges         bool
	repoLayout         env_repo.Env
	preWrite           interfaces.FuncIter[*sku.Transacted]
	config             store_config.Store
	oids               map[string]struct{}
}

func (pt *Page) initialize(
	pid PageId,
	i *Index,
) {
	pt.repoLayout = i.directoryLayout
	pt.sunrise = i.sunrise
	pt.PageId = pid
	pt.added = sku.MakeList()
	pt.addedLatest = sku.MakeList()
	pt.preWrite = i.preWrite
	pt.probe_index = &i.probe_index
	pt.oids = make(map[string]struct{})
}

func (s *Page) readOneRange(
	ra object_probe_index.Range,
	sk *sku.Transacted,
) (err error) {
	var f *os.File

	if f, err = files.Open(s.Path()); err != nil {
		err = errors.Wrap(err)
		return
	}

	defer errors.DeferredCloser(&err, f)

	b := make([]byte, ra.ContentLength)

	if _, err = f.ReadAt(b, ra.Offset); err != nil {
		err = errors.Wrapf(err, "Range: %q, Page: %q", ra, s.PageId)
		return
	}

	dec := makeBinaryWithQueryGroup(nil, ids.SigilHistory)

	skWR := skuWithRangeAndSigil{
		skuWithSigil: skuWithSigil{
			Transacted: sk,
		},
		Range: ra,
	}

	if _, err = dec.readFormatExactly(f, &skWR); err != nil {
		err = errors.Wrapf(err, "Range: %q, Page: %q", ra, s.PageId.Path())
		return
	}

	return
}

func (pt *Page) add(
	z1 *sku.Transacted,
	options sku.CommitOptions,
) (err error) {
	pt.oids[z1.ObjectId.String()] = struct{}{}
	z := sku.GetTransactedPool().Get()

	sku.TransactedResetter.ResetWith(z, z1)

	if pt.sunrise.Less(z.GetTai()) || options.StreamIndexOptions.ForceLatest {
		pt.addedLatest.Add(z)
	} else {
		pt.added.Add(z)
	}

	pt.hasChanges = true

	return
}

func (pt *Page) waitingToAddLen() int {
	return pt.added.Len() + pt.addedLatest.Len()
}

func (pt *Page) copyJustHistoryFrom(
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

func (pt *Page) copyJustHistoryAndAdded(
	s sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
) (err error) {
	return pt.copyHistoryAndMaybeLatest(s, w, true, false)
}

func (pt *Page) copyHistoryAndMaybeLatest(
	qg sku.PrimitiveQueryGroup,
	w interfaces.FuncIter[*sku.Transacted],
	includeAdded bool,
	includeAddedLatest bool,
) (err error) {
	var r io.ReadCloser

	if r, err = pt.repoLayout.ReadCloserCache(pt.Path()); err != nil {
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
		if err = pt.copyJustHistoryFrom(
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
	added := pt.added.Copy()

	var sk skuWithRangeAndSigil

	if err = heap.MergeStream(
		&added,
		func() (tz *sku.Transacted, err error) {
			tz = sku.GetTransactedPool().Get()
			sk.Transacted = tz
			_, err = dec.readFormatAndMatchSigil(br, &sk)
			if err != nil {
				if errors.IsEOF(err) {
					err = collections.MakeErrStopIteration()
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

	addedLatest := pt.addedLatest.Copy()

	if err = heap.MergeStream(
		&addedLatest,
		func() (tz *sku.Transacted, err error) {
			err = collections.MakeErrStopIteration()
			return
		},
		w,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (pt *Page) MakeFlush(
	changesAreHistorical bool,
) func() error {
	return func() (err error) {
		pw := &writer{
			Page:        pt,
			probe_index: pt.probe_index,
		}

		if changesAreHistorical {
			pw.changesAreHistorical = true
			pw.hasChanges = true
		}

		if err = pw.Flush(); err != nil {
			err = errors.Wrap(err)
			return
		}

		pt.hasChanges = false

		return
	}
}
