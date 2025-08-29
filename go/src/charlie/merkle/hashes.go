package merkle

import (
	"crypto/sha256"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

type Hash struct {
	hash.Hash
	tipe string
}

func (hash Hash) GetType() string {
	return hash.tipe
}

var HashTypeSha256 = HashType{
	pool: pool.MakeValue(
		func() Hash {
			return Hash{
				Hash: sha256.New(),
				tipe: HRPObjectBlobDigestSha256V0,
			}
		},
		func(hash Hash) {
			hash.Reset()
		},
	),
	tipe: HRPObjectBlobDigestSha256V0,
}

type HashType struct {
	pool interfaces.PoolValue[Hash]
	tipe string
}

func (hashType HashType) Get() Hash {
	return hashType.pool.Get()
}

// TODO assert type somehow?
func (hashType HashType) Put(hash Hash) {
	hashType.pool.Put(hash)
}

func (hashType HashType) FromStringContent(input string) interfaces.BlobId {
	stringReader, repoolStringReader := pool.GetStringReader(input)
	defer repoolStringReader()

	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	if _, err := io.Copy(hash, stringReader); err != nil {
		errors.PanicIfError(err)
	}

	digest := hash.Sum(nil)

	var id Id

	errors.PanicIfError(id.SetMerkleId(hash.tipe, digest))

	return id
}
