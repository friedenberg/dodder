package merkle

import (
	"crypto/sha256"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

type HashType struct {
	pool interfaces.PoolValue[hash.Hash]
	tipe string
}

var HashTypeSha256 = HashType{
	pool: pool.MakeValue(
		func() hash.Hash {
			return sha256.New()
		},
		func(hash hash.Hash) {
			hash.Reset()
		},
	),
	tipe: HRPObjectBlobDigestSha256V0,
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

	errors.PanicIfError(id.SetMerkleId(hashType.tipe, digest))

	return id
}
