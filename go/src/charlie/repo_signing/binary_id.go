package repo_signing

import (
	"slices"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var _ interfaces.MutableGenericBlobId = &BinaryId{}

type BinaryId struct {
	tipe string // hrp
	data []byte
}

func (binaryId BinaryId) IsEmpty() bool {
	return len(binaryId.data) == 0
}

func (binaryId BinaryId) GetSize() int {
	return len(binaryId.data)
}

func (binaryId BinaryId) GetBytes() []byte {
	return binaryId.data
}

func (binaryId BinaryId) GetType() string {
	return binaryId.tipe
}

func (binaryId BinaryId) IsNull() bool {
	return len(binaryId.data) == 0
}

func (binaryId *BinaryId) SetDigest(digest interfaces.BlobId) error {
	return errors.Err501NotImplemented
}

func (binaryId *BinaryId) SetType(tipe string) (err error) {
	binaryId.tipe = tipe
	return
}

func (binaryId *BinaryId) SetBytes(bytes []byte) error {
	binaryId.data = slices.Grow(binaryId.data, len(bytes))
	binaryId.data = binaryId.data[:cap(binaryId.data)]
	copy(binaryId.data, bytes)
	return nil
}

func (binaryId *BinaryId) Reset() {
	binaryId.tipe = ""
	binaryId.data = binaryId.data[:0]
}

func (binaryId *BinaryId) GetBlobId() interfaces.BlobId {
	return binaryId
}
