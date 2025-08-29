package merkle

import (
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type Hash struct {
	hash    hash.Hash
	tipe    string
	written int64
}

var _ interfaces.Hash = &Hash{}

func (hash *Hash) Write(bites []byte) (written int, err error) {
	written, err = hash.hash.Write(bites)
	hash.written += int64(written)
	return written, err
}

func (hash *Hash) Sum(bites []byte) []byte {
	return hash.hash.Sum(bites)
}

func (hash *Hash) Reset() {
	hash.written = 0
	hash.hash.Reset()
}

func (hash *Hash) Size() int {
	return hash.hash.Size()
}

func (hash *Hash) BlockSize() int {
	return hash.hash.BlockSize()
}

func (hash *Hash) GetType() string {
	return hash.tipe
}

func (hash *Hash) GetBlobId() (interfaces.BlobId, interfaces.FuncRepool) {
	id := idPool.Get()

	var digestBytes []byte

	if hash.written > 0 {
		// TODO verify this works as expected
		digestBytes = hash.hash.Sum(id.data)
	}

	errors.PanicIfError(id.SetMerkleId(hash.tipe, digestBytes))

	return id, func() {
		idPool.Put(id)
	}
}
