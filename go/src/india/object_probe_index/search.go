package object_probe_index

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/delta/sha"
)

func (page *page) seekToFirstBinarySearch(
	shMet interfaces.Digest,
) (mid int64, err error) {
	if page.file == nil {
		err = collections.MakeErrNotFoundString(
			"fd nil: " + digests.Format(shMet),
		)
		return
	}

	var low, hi int64
	shMid := &sha.Sha{}

	var rowCount int64

	if rowCount, err = page.GetRowCount(); err != nil {
		err = errors.Wrap(err)
		return
	}

	hi = rowCount
	loops := 0

	for low <= hi {
		loops++
		mid = (hi + low) / 2

		// var loc int64

		if _, err = shMid.ReadAtFrom(page.file, mid*RowSize); err != nil {
			err = errors.Wrap(err)
			return
		}

		cmp := bytes.Compare(shMet.GetBytes(), shMid.GetBytes())

		switch cmp {
		case -1:
			if low == hi-1 {
				low = hi
			} else {
				hi = mid - 1
			}

		case 0:
			// found
			return

		case 1:
			low = mid + 1

		default:
			panic("not possible")
		}
	}

	err = collections.MakeErrNotFoundString(
		fmt.Sprintf("%d: %s", loops, digests.Format(shMet)),
	)

	return
}

func (page *page) seekToFirstLinearSearch(
	shMet interfaces.Digest,
) (loc int64, err error) {
	if page.file == nil {
		err = collections.MakeErrNotFoundString(
			"fd nil: " + digests.Format(shMet),
		)
		return
	}

	var rowCount int64
	shMid := &sha.Sha{}

	if rowCount, err = page.GetRowCount(); err != nil {
		err = errors.Wrap(err)
		return
	}

	page.bufferedReader.Reset(page.file)
	buf := bytes.NewBuffer(make([]byte, RowSize))
	buf.Reset()

	for loc = int64(0); loc <= rowCount; loc++ {
		// var loc int64

		if _, err = buf.ReadFrom(&page.bufferedReader); err != nil {
			err = errors.Wrap(err)
			return
		}

		if _, err = shMid.ReadFrom(buf); err != nil {
			err = errors.Wrap(err)
			return
		}

		if bytes.Equal(shMet.GetBytes(), shMid.GetBytes()) {
			// found
			return
		}
	}

	err = collections.MakeErrNotFoundString(digests.Format(shMet))

	return
}
