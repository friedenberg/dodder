package repo_signing

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

var _ interfaces.MutableGenericBlobId = &BinaryId{}

type BinaryId struct {
	tipe string // hrp
	data []byte
}

func (binaryId BinaryId) String() string {
	return fmt.Sprintf("%s-%x", binaryId.tipe, binaryId.data)
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

func (binaryId *BinaryId) SetDigest(digest interfaces.BlobId) (err error) {
	if err = binaryId.SetType(digest.GetType()); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = binaryId.SetBytes(digest.GetBytes()); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (binaryId *BinaryId) SetType(tipe string) (err error) {
	if tipe == "" {
		err = errors.Errorf("empty type")
		return
	}

	binaryId.tipe = tipe

	return
}

// TODO optimize this
func (binaryId *BinaryId) SetBytes(bytes []byte) error {
	binaryId.data = make([]byte, len(bytes))
	// binaryId.data = slices.Grow(binaryId.data, len(bytes)-len(binaryId.data))
	// binaryId.data = binaryId.data[:cap(binaryId.data)]
	copy(binaryId.data, bytes)
	return nil
}

func (binaryId *BinaryId) Reset() {
	binaryId.tipe = ""
	binaryId.data = binaryId.data[:0]
}

func (binaryId *BinaryId) ResetWith(src *BinaryId) {
	binaryId.tipe = src.tipe
	errors.PanicIfError(binaryId.SetBytes(src.GetBytes()))
}

func (binaryId *BinaryId) GetBlobId() interfaces.BlobId {
	return binaryId
}
