package merkle

import (
	"crypto/sha256"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var HashTypeSha256 = HashType{
	pool: pool.MakeValue(
		func() Hash {
			return Hash{
				hash: sha256.New(),
				tipe: HRPObjectBlobDigestSha256V0,
			}
		},
		func(hash Hash) {
			hash.hash.Reset()
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

func (hashType HashType) Put(hash Hash) {
	errors.PanicIfError(MakeErrWrongType(hashType.tipe, hash.tipe))
	hashType.pool.Put(hash)
}

func (hashType HashType) FromStringContent(input string) interfaces.BlobId {
	stringReader, repoolStringReader := pool.GetStringReader(input)
	defer repoolStringReader()

	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	if _, err := io.Copy(hash.hash, stringReader); err != nil {
		errors.PanicIfError(err)
	}

	id, _ := hash.GetBlobId()

	return id
}
