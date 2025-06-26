package id

import (
	"os"
	"path"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type TypedId interface {
	interfaces.Genre
	interfaces.Setter
}

func Path(i interfaces.StringerWithHeadAndTail, pc ...string) string {
	pc = append(pc, i.GetHead(), i.GetTail())
	return path.Join(pc...)
}

func MakeDirIfNecessary(i interfaces.StringerWithHeadAndTail, pc ...string) (p string, err error) {
	p = Path(i, pc...)
	dir := path.Dir(p)

	if err = os.MkdirAll(dir, os.ModeDir|0o755); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
