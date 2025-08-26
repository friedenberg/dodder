package key_bytes

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
	ContentLength = Binary('C')
	Sigil         = Binary('S')
	Blob          = Binary('A')
	RepoPubKey    = Binary('P')
	RepoSig       = Binary('q')
	Description   = Binary('B')
	Tag           = Binary('E')
	Genre         = Binary('G')
	ObjectId      = Binary('K')
	Comment       = Binary('k')
	Tai           = Binary('T')
	Type          = Binary('t')

	SigParentMetadataParentObjectId = Binary('M')
	DigestMetadataParentObjectId    = Binary('s')
	DigestMetadataWithoutTai        = Binary('n')
	DigestMetadata                  = Binary('m')

	CacheParentTai   = Binary('p')
	CacheDormant     = Binary('a')
	CacheTagImplicit = Binary('I')
	CacheTagExpanded = Binary('e')
	CacheTags        = Binary('x')
	CacheTags2       = Binary('y')
)

var ErrInvalid = errors.New("invalid key")

func (key Binary) String() string {
	return fmt.Sprintf("%c", byte(key))
}

func (key *Binary) Reset() {
	*key = 0
}

func (key Binary) ReadByte() (byte, error) {
	return byte(key), nil
}

func (key Binary) WriteTo(w io.Writer) (n int64, err error) {
	bites := [1]byte{byte(key)}
	var n1 int
	n1, err = ohio.WriteAllOrDieTrying(w, bites[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	return
}

func (key *Binary) WriteByte(b byte) (err error) {
	*key = Binary(b)

	return
}

func (key *Binary) ReadFrom(r io.Reader) (n int64, err error) {
	var bite [1]byte
	var n1 int
	n1, err = ohio.ReadAllOrDieTrying(r, bite[:])
	n += int64(n1)

	if err != nil {
		err = errors.WrapExceptSentinel(err, io.EOF)
		return
	}

	err = key.WriteByte(bite[0])

	return
}
