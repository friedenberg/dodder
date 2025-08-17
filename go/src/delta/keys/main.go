package keys

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

type Binary byte

const (
	// TODO make this less fragile by guaranteeing unique values
	Unknown       = Binary(iota)
	ContentLength = 'C'
	Sigil         = 'S'
	Blob          = 'A'
	RepoPubKey    = 'P'
	RepoSig       = 'q'
	Description   = 'B'
	Tag           = 'E'
	Genre         = 'G'
	ObjectId      = 'K'
	Comment       = 'k'
	Tai           = 'T'
	Type          = 't'

	DigestParentMetadataParentObjectId = 'M'
	DigestMetadataParentObjectId       = 's'
	DigestMetadataWithoutTai           = 'n'
	DigestMetadata                     = 'm'

	CacheParentTai   = 'p'
	CacheDormant     = 'a'
	CacheTagImplicit = 'I'
	CacheTagExpanded = 'e'
	CacheTags        = 'x'
	CacheTags2       = 'y'
)

var ErrInvalid = errors.New("invalid key")

func (s Binary) String() string {
	return fmt.Sprintf("%c", byte(s))
}

func (s *Binary) Reset() {
	*s = 0
}

func (s *Binary) ReadByte() (byte, error) {
	return byte(*s), nil
}

func (s *Binary) WriteTo(w io.Writer) (n int64, err error) {
	b := [1]byte{byte(*s)}
	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, b[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (s *Binary) WriteByte(b byte) (err error) {
	*s = Binary(b)

	return
}

func (s *Binary) ReadFrom(r io.Reader) (n int64, err error) {
	var b [1]byte
	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, b[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	err = s.WriteByte(b[0])

	return
}
