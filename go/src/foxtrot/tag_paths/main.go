package tag_paths

import (
	"bytes"
	"io"
	"sort"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
	"code.linenisgreat.com/dodder/go/src/delta/catgut"
)

type (
	Tag  = catgut.String
	Path []*Tag
)

func MakePathWithType(tags ...*Tag) *PathWithType {
	return &PathWithType{
		Path: makePath(tags...),
	}
}

func makePath(tags ...*Tag) Path {
	path := Path(make([]*Tag, 0, len(tags)))

	for _, e := range tags {
		path.Add(e)
	}

	return path
}

func (path *Path) Clone() *Path {
	clone := makePath(*path...)
	return &clone
}

func (path *Path) CloneAndAddPath(c *Path) *Path {
	var clone Path
	if path == nil {
		clone = makePath()
	} else {
		clone = makePath(*path...)
	}

	clone.AddPath(c)

	return &clone
}

func (path *Path) IsEmpty() bool {
	if path == nil {
		return true
	}

	return path.Len() == 0
}

func (path *Path) First() *Tag {
	return (*path)[0]
}

func (path *Path) Last() *Tag {
	return (*path)[path.Len()-1]
}

func (path *Path) Equals(b *Path) bool {
	if path.Len() != b.Len() {
		return false
	}

	for i, as := range *path {
		if !as.Equals((*b)[i]) {
			return false
		}
	}

	return true
}

func (path *Path) Compare(b *Path) int {
	elsA := *path
	elsB := *b

	for {
		lenA, lenB := len(elsA), len(elsB)

		switch {
		case lenA == 0 && lenB == 0:
			return 0

		case lenA == 0:
			return -1

		case lenB == 0:
			return 1
		}

		elA := elsA[0]
		elsA = elsA[1:]

		elB := elsB[0]
		elsB = elsB[1:]

		cmp := elA.Compare(elB)

		if cmp != 0 {
			return cmp
		}
	}

	return 0
}

func (path *Path) String() string {
	return (*StringBackward)(path).String()
}

func (path *Path) Copy() (b *Path) {
	b = &Path{}
	*b = make([]*Tag, path.Len())

	if path == nil {
		return b
	}

	for i, s := range *path {
		sb := catgut.GetPool().Get()
		s.CopyTo(sb)
		(*b)[i] = sb
	}

	return b
}

func (path *Path) Len() int {
	if path == nil {
		return 0
	}

	return len(*path)
}

func (path *Path) Cap() int {
	if path == nil {
		return 0
	}

	return cap(*path)
}

func (path *Path) Less(i, j int) bool {
	return bytes.Compare((*path)[i].Bytes(), (*path)[i].Bytes()) == -1
}

func (path *Path) Swap(i, j int) {
	a, b := (*path)[i], (*path)[j]
	var x Tag
	x.SetBytes(a.Bytes())
	a.SetBytes(b.Bytes())
	b.SetBytes(x.Bytes())
}

func (path *Path) AddPath(b *Path) {
	if b.IsEmpty() {
		return
	}

	for _, e := range *b {
		*path = append(*path, catgut.GetPool().Get())
		(*path)[path.Len()-1].SetBytes(e.Bytes())
	}

	sort.Sort(path)
}

func (path *Path) Add(es ...*Tag) {
	for _, e := range es {
		if e.IsEmpty() {
			return
		}

		if path.Len() > 0 && (*path)[path.Len()-1].Compare(e) == 0 {
			return
		}

		*path = append(*path, catgut.GetPool().Get())
		(*path)[path.Len()-1].SetBytes(e.Bytes())
	}

	sort.Sort(path)
}

func (path *Path) ReadFrom(r io.Reader) (n int64, err error) {
	var count uint8

	var n1 int
	if count, n1, err = ohio.ReadUint8(r); err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return n, err
	}

	n += int64(n1)

	*path = (*path)[:path.Cap()]

	if diff := count - uint8(path.Len()); diff > 0 {
		*path = append(*path, make([]*Tag, diff)...)
	}

	for i := uint8(0); i < count; i++ {
		var cl uint8

		if cl, n1, err = ohio.ReadUint8(r); err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return n, err
		}

		n += int64(n1)

		if (*path)[i] == nil {
			(*path)[i] = catgut.GetPool().Get()
		}

		_, err = (*path)[i].ReadNFrom(r, int(cl))
		if err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return n, err
		}
	}

	return n, err
}

func (path *Path) WriteTo(w io.Writer) (n int64, err error) {
	var n1 int

	n1, err = ohio.WriteUint8(w, uint8(path.Len()))
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return n, err
	}

	for _, s := range *path {
		if s.Len() == 0 {
			panic("found empty tag in tag_paths")
		}

		n1, err = ohio.WriteUint8(w, uint8(s.Len()))
		n += int64(n1)

		if err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return n, err
		}

		var n2 int64
		n2, err = s.WriteTo(w)
		n += n2

		if err != nil {
			err = errors.WrapExceptSentinel(err, io.EOF)
			return n, err
		}
	}

	return n, err
}
