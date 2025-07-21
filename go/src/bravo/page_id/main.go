package page_id

import (
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/digests"
)

// TODO move to own package
type PageId struct {
	Index  uint8
	Dir    string
	Prefix string
}

func PageIdFromPath(n uint8, p string) PageId {
	dir, file := filepath.Split(p)
	return PageId{
		Dir:    dir,
		Prefix: file,
		Index:  n,
	}
}

func (pid PageId) String() string {
	return fmt.Sprintf("%d", pid.Index)
}

func (pid *PageId) Path() string {
	return filepath.Join(pid.Dir, fmt.Sprintf("%s-%x", pid.Prefix, pid.Index))
}

func PageIndexForString(
	width uint8,
	value string,
	digester interfaces.WriteDigester,
) (n uint8, err error) {
	stringReader := strings.NewReader(value)

	if _, err = io.Copy(digester, stringReader); err != nil {
		err = errors.Wrap(err)
		return
	}

	// TODO figure out how to repool these
	digest := digester.GetDigest()

	if n, err = PageIndexForDigest(width, digest); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForDigest(
	width uint8,
	digest interfaces.Digest,
) (n uint8, err error) {
	var n1 int64

	bucketIndexString := digests.FormatDigest(digest)[:width]

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
