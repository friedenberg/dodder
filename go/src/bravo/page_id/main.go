package page_id

import (
	"fmt"
	"math"
	"path/filepath"
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/markl"
)

// TODO rename to Id
type PageId struct {
	Index  uint8
	Dir    string
	Prefix string
}

func PageIdFromPath(n uint8, path string) PageId {
	dir, file := filepath.Split(path)
	return PageId{
		Dir:    dir,
		Prefix: file,
		Index:  n,
	}
}

func (id PageId) String() string {
	return fmt.Sprintf("%d", id.Index)
}

func (id *PageId) Path() string {
	return filepath.Join(id.Dir, fmt.Sprintf("%s-%x", id.Prefix, id.Index))
}

func PageIndexForString(
	width uint8,
	value string,
	hashType interfaces.HashType,
) (n uint8, err error) {
	digest, repool := hashType.GetMarklIdForString(value)
	defer repool()

	if n, err = PageIndexForDigest(width, digest); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForDigest(
	width uint8,
	digest interfaces.MarklId,
) (n uint8, err error) {
	var n1 int64

	if err = markl.MakeErrIsNull(digest, "page id"); err != nil {
		panic(err)
	}

	digestString := markl.Format(digest)
	bucketIndexString := digestString[:width]

	if n1, err = strconv.ParseInt(bucketIndexString, 16, 64); err != nil {
		err = errors.Wrap(err)
		return
	}

	if n1 > math.MaxUint8 {
		err = errors.ErrorWithStackf("page out of bounds: %d", n1)
		return
	}

	n = uint8(n1)

	return
}
