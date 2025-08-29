package merkle

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

func FromStringContent(hashType HashType, input string) interfaces.BlobId {
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
