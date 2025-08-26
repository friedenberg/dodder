package merkle

import (
	"bytes"
	"fmt"
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

var (
	_ interfaces.BlobId        = Id{}
	_ interfaces.MutableBlobId = &Id{}
)

type Id struct {
	tipe string
	data []byte
}

func (id Id) String() string {
	if id.tipe == "" && len(id.data) == 0 {
		return ""
	}

	if id.tipe == HRPObjectBlobDigestSha256V0 {
		return fmt.Sprintf("%x", id.data)
	} else {
		bites, err := blech32.Encode(id.tipe, id.data)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func (id Id) IsEmpty() bool {
	return len(id.data) == 0
}

func (id Id) GetSize() int {
	return len(id.data)
}

func (id Id) GetBytes() []byte {
	return id.data
}

func (id Id) GetType() string {
	return id.tipe
}

func (id Id) IsNull() bool {
	return len(id.data) == 0
}

func (id *Id) Set(value string) (err error) {
	if id.tipe, id.data, err = blech32.DecodeString(value); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMerkleId(id.tipe, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetDigest(digest interfaces.BlobId) (err error) {
	if err = id.SetMerkleId(digest.GetType(), digest.GetBytes()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) SetMerkleId(tipe string, bites []byte) (err error) {
	if tipe == "" && len(bites) == 0 {
		id.Reset()
		return
	}

	if !slices.Contains(hrpValid, tipe) {
		err = errors.Errorf("invalid type: %q", tipe)
		return
	}

	id.tipe = tipe

	// TODO optimize this and validate against type
	id.data = make([]byte, len(bites))
	// binaryId.data = slices.Grow(binaryId.data, len(bytes)-len(binaryId.data))
	// binaryId.data = binaryId.data[:cap(binaryId.data)]
	copy(id.data, bites)

	return
}

func (id *Id) Reset() {
	id.tipe = ""
	id.data = id.data[:0]
}

func (id *Id) ResetWith(src *Id) {
	id.tipe = src.tipe
	bites := src.data
	id.data = make([]byte, len(bites))
	// binaryId.data = slices.Grow(binaryId.data, len(bytes)-len(binaryId.data))
	// binaryId.data = binaryId.data[:cap(binaryId.data)]
	copy(id.data, bites)
}

func (id *Id) ResetWithMerkleId(src interfaces.BlobId) {
	errors.PanicIfError(id.SetMerkleId(src.GetType(), src.GetBytes()))
}

func (id *Id) GetBlobId() interfaces.BlobId {
	return id
}

func (id *Id) UnmarshalBinary(
	bites []byte,
) (err error) {
	if len(bites) == 0 {
		return
	}

	tipeBytes, bytesAfterTipe, ok := bytes.Cut(bites, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = id.SetMerkleId(string(tipeBytes), bytesAfterTipe); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// structure (in bytes):
// <256: type
// 1: empty byte
// <256: id
func (id Id) MarshalBinary() (bytes []byte, err error) {
	// TODO confirm few allocations
	// TODO confirm size of type is less than 256
	tipe := id.GetType()
	bites := id.GetBytes()

	if tipe == "" && len(bites) == 0 {
		return
	} else if tipe == "" {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(tipe)...)
	bytes = append(bytes, '\x00')
	bytes = append(bytes, bites...)

	return
}

func (id Id) MarshalText() (bites []byte, err error) {
	if id.tipe == HRPObjectBlobDigestSha256V0 {
		bites = fmt.Appendf(nil, "%x", id.data)
	} else {
		if bites, err = blech32.Encode(id.tipe, id.data); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (id *Id) UnmarshalText(bites []byte) (err error) {
	if id.tipe, id.data, err = blech32.Decode(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
