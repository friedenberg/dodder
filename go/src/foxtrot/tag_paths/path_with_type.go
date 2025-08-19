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

func (path *PathWithType) String() string {
	return fmt.Sprintf(
		"%s:%s",
		path.Type.String(),
		(*StringBackward)(&path.Path).String(),
	)
}

func (path *PathWithType) Clone() (clone *PathWithType) {
	clone = MakePathWithType(path.Path...)
	clone.Type = path.Type

	return
}

func (path *PathWithType) ReadFrom(r io.Reader) (n int64, err error) {
	var n1 int64
	n1, err = path.Type.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = path.Path.ReadFrom(r)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (path *PathWithType) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int64
	n1, err = path.Type.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	n1, err = path.Path.WriteTo(w)
	n += n1

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
