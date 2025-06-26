package tag_paths

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

//go:generate stringer -type=Type
type Type byte

// describe these
const (
	TypeDirect = Type(iota)
	TypeSuper
	TypeIndirect
	TypeSelf
	TypeUnknown
)

// TODO determine if this should include type self
func (t Type) IsDirectOrSelf() bool {
	switch t {
	case TypeDirect, TypeSelf:
		return true

	default:
		return false
	}
}

func (t *Type) SetDirect() {
	*t = TypeDirect
}

func (t Type) ReadByte() (byte, error) {
	return byte(t), nil
}

func (t *Type) ReadFrom(r io.Reader) (n int64, err error) {
	var b [1]byte

	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b[:])
	n = int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	*t = Type(b[0])

	return
}

func (t Type) WriteTo(w io.Writer) (n int64, err error) {
	var b byte

	if b, err = t.ReadByte(); err != nil {
		err = errors.Wrap(err)
		return
	}

	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, []byte{b})
	n = int64(n1)

	if err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
