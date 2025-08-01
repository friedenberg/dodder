package sha

import (
	"bufio"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO move to generic digest package
type Slice []*Sha

func (slice *Slice) ReadFrom(reader io.Reader) (n int64, err error) {
	// TODO use pool
	bufferedReader := bufio.NewReader(reader)

	var isEOF bool

	for !isEOF {
		var line string
		line, err = bufferedReader.ReadString('\n')

		if err == io.EOF {
			err = nil
			isEOF = true
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		if line == "" {
			continue
		}

		sh := poolSha.Get()

		if err = sh.Set(strings.TrimSpace(line)); err != nil {
			err = errors.Wrap(err)
			return
		}

		*slice = append(*slice, sh)
	}

	return
}
