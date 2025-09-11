package object_probe_index

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/collections"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

func (page *page) seekToFirstBinarySearch(
	expected interfaces.MarklId,
) (mid int64, err error) {
	errors.PanicIfError(
		markl.MakeErrWrongType(
			page.hashType.GetMarklTypeId(),
			expected.GetMarklType().GetMarklTypeId(),
		),
	)

	if page.file == nil {
		err = collections.MakeErrNotFoundString(
			"fd nil: " + markl.FormatBytesAsHext(expected),
		)
		return
	}

	var low, hi int64

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

		cmp := markl.CompareToReaderAt(
			page.file,
			mid*int64(page.rowWidth),
			expected,
		)

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
		fmt.Sprintf("%d: %s", loops, markl.FormatBytesAsHext(expected)),
	)

	return
}

func (page *page) seekToFirstLinearSearch(
	expected interfaces.MarklId,
) (loc int64, err error) {
	errors.PanicIfError(
		markl.MakeErrWrongType(
			page.hashType.GetMarklTypeId(),
			expected.GetMarklType().GetMarklTypeId(),
		),
	)

	if page.file == nil {
		err = collections.MakeErrNotFoundString(
			"fd nil: " + markl.FormatBytesAsHext(expected),
		)
		return
	}

	var rowCount int64

	if rowCount, err = page.GetRowCount(); err != nil {
		err = errors.Wrap(err)
		return
	}

	page.bufferedReader.Reset(page.file)

	for loc = int64(0); loc <= rowCount; loc++ {
		// var loc int64

		if markl.CompareToReader(&page.bufferedReader, expected) == 0 {
			return
		}
	}

	err = collections.MakeErrNotFoundString(markl.FormatBytesAsHext(expected))

	return
}
