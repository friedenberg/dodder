package blob_ids

import (
	"bytes"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

// func UnmarshalBinary(
// 	blobId interfaces.MutableBlobId,
// 	bytes []byte,
// ) (err error) {
// 	if blobId, ok := blobId.(interfaces.MutableGenericBlobId); ok {
// 		return UnmarshalBinaryGeneric(blobId, bytes)
// 	}

// 	return errors.Err501NotImplemented
// }

type BlobIdBinaryUnmarshaler struct {
	interfaces.MutableGenericBlobId
}

func (blobId BlobIdBinaryUnmarshaler) UnmarshalBinary(
	bites []byte,
) (err error) {
	tipeBytes, bytesAfterTipe, ok := bytes.Cut(bites, []byte{'\x00'})

	if !ok {
		err = errors.Errorf("expected empty byte, but none found")
		return
	}

	if err = blobId.SetType(string(tipeBytes)); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = blobId.SetBytes(bytesAfterTipe); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

type BlobIdBinaryMarshaler struct {
	interfaces.BlobId
}

// structure (in bytes):
// <256: type
// 1: empty byte
// <256: id
func (blobId BlobIdBinaryMarshaler) MarshalBinary() (bytes []byte, err error) {
	// TODO confirm few allocations
	// TODO confirm size of type is less than 256
	tipe := blobId.GetType()

	if tipe == "" {
		err = errors.Errorf("empty type")
		return
	}

	bytes = append(bytes, []byte(tipe)...)
	bytes = append(bytes, '\x00')
	idBytes := blobId.GetBytes()
	bytes = append(bytes, idBytes...)

	return
}
