package tag_paths

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type PathWithType struct {
	Path
	Type
}

func (p *PathWithType) String() string {
	return fmt.Sprintf(
		"%s:%s",
		p.Type.String(),
		(*StringBackward)(&p.Path).String(),
	)
}

func (a *PathWithType) Clone() (b *PathWithType) {
	b = MakePathWithType(a.Path...)
	b.Type = a.Type

	return
}

func (p *PathWithType) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = p.Type.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = p.Path.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (p *PathWithType) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int64
	n1, err = p.Type.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = p.Path.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
