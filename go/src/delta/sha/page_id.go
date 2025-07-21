package sha

import (
	"crypto/sha256"
	"fmt"
	"io"
	"math"
	"path/filepath"
	"strconv"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

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

func PageIndexForString(width uint8, s string) (n uint8, err error) {
	sr := strings.NewReader(s)
	hash := sha256.New()

	if _, err = io.Copy(hash, sr); err != nil {
		err = errors.Wrap(err)
		return
	}

	sh := FromHash(hash)
	defer GetPool().Put(sh)

	if n, err = PageIndexForSha(width, sh); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func PageIndexForSha(width uint8, digest interfaces.Digest) (n uint8, err error) {
	var n1 int64
	ss := interfaces.FormatDigest(digest)[:width]

	if n1, err = strconv.ParseInt(ss, 16, 64); err != nil {
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
