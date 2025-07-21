package sha

import (
	"bufio"
	"io"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO move to generic digest package
type Slice []*Sha

func (s *Slice) ReadFrom(r io.Reader) (n int64, err error) {
	br := bufio.NewReader(r)

	var eof bool

	for !eof {
		var line string
		line, err = br.ReadString('\n')

		if err == io.EOF {
			err = nil
			eof = true
		} else if err != nil {
			err = errors.Wrap(err)
			return
		}

		if line == "" {
			continue
		}

		sh := GetPool().Get()

		if err = sh.Set(strings.TrimSpace(line)); err != nil {
			err = errors.Wrap(err)
			return
		}

		*s = append(*s, sh)
	}

	return
}
