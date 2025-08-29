package merkle

import (
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Hash struct {
	hash hash.Hash
	tipe string
}

var _ interfaces.Hash = Hash{}

func (hash Hash) Write(bites []byte) (int, error) {
	return hash.hash.Write(bites)
}

func (hash Hash) Sum(bites []byte) []byte {
	return hash.hash.Sum(bites)
}

func (hash Hash) Reset() {
	hash.hash.Reset()
}

func (hash Hash) Size() int {
	return hash.hash.Size()
}

func (hash Hash) BlockSize() int {
	return hash.hash.BlockSize()
}

func (hash Hash) GetType() string {
	return hash.tipe
}

func (hash Hash) GetBlobId() (interfaces.BlobId, interfaces.FuncRepool) {
	id := idPool.Get()

	// TODO verify this works as expected
	digest := hash.hash.Sum(id.data)

	errors.PanicIfError(id.SetMerkleId(hash.tipe, digest))

	return id, func() {
		idPool.Put(id)
	}
}
