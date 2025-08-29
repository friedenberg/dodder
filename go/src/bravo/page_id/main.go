package page_id

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"code.linenisgreat.com/dodder/go/src/charlie/merkle"
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
	envDigest interfaces.EnvBlobId,
) (n uint8, err error) {
	stringReader, repool := pool.GetStringReader(value)
	defer repool()

	writer, repool := envDigest.MakeWriteDigesterWithRepool()
	defer repool()

	if _, err = io.Copy(writer, stringReader); err != nil {
		err = errors.Wrap(err)
		return
	}

	digest := envDigest.GetBlobId()
	defer envDigest.PutBlobId(digest)

	if n, err = PageIndexForDigest(width, digest); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForDigest(
	width uint8,
	digest interfaces.BlobId,
) (n uint8, err error) {
	var n1 int64

	bucketIndexString := merkle.Format(digest)[:width]

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
