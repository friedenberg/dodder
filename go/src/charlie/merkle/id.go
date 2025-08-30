package merkle

import (
	"bytes"
	"encoding/hex"
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
	if len(id.data) == 0 {
		return true
	} else if id.tipe == "" {
		panic("empty type")
	}

	hashType, ok := hrpToHashType[id.tipe]

	// this is not an Id for a hash, so it can never be null with non-zero data
	// contents
	if !ok {
		return false
	}

	if bytes.Equal(id.data, hashType.null.GetBytes()) {
		return true
	}

	return false
}

func (id *Id) SetMaybeSha256(value string) (err error) {
	if len(value) == 64 {
		if err = id.SetSha256(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		if err = id.Set(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (id *Id) SetSha256(value string) (err error) {
	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "%q", value)
		return
	}

	if err = id.SetMerkleId(
		HRPObjectBlobDigestSha256V0,
		decodedBytes,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) Set(value string) (err error) {
	if id.tipe, id.data, err = blech32.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "Value: %q", value)
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
	id.setData(bites)

	return
}

func (id *Id) allocDataIfNecessary(size int) {
	id.data = id.data[:0]
	id.data = slices.Grow(id.data, size)
}

func (id *Id) allocDataAndSetToCapIfNecessary(size int) {
	id.allocDataIfNecessary(size)
	id.data = id.data[:size]
}

func (id *Id) setData(data []byte) {
	// TODO validate against type
	id.allocDataAndSetToCapIfNecessary(len(data))
	copy(id.data, data)
}

func (id *Id) Reset() {
	id.tipe = ""
	id.data = id.data[:0]
}

func (id *Id) ResetWith(src Id) {
	id.tipe = src.tipe
	id.setData(src.data)
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
