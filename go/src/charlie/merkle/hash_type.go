package merkle

import (
	"crypto"
	"crypto/sha256"
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

var HashTypeSha256 = HashType{
	pool: pool.MakeValue(
		func() interfaces.Hash {
			return &Hash{
				hash: sha256.New(),
				tipe: HRPObjectBlobDigestSha256V0,
			}
		},
		func(hash interfaces.Hash) {
			hash.Reset()
		},
	),
	tipe:  HRPObjectBlobDigestSha256V0,
	width: 32,
}

type HashType struct {
	crypto.Hash
	pool  interfaces.PoolValue[interfaces.Hash]
	tipe  string
	width int
}

var _ interfaces.HashType = HashType{}

func (hashType HashType) Get() interfaces.Hash {
	return hashType.pool.Get()
}

func (hashType HashType) Put(hash interfaces.Hash) {
	errors.PanicIfError(MakeErrWrongType(hashType.tipe, hash.GetType()))
	hashType.pool.Put(hash)
}

func (hashType HashType) GetType() string {
	return hashType.tipe
}

func (hashType HashType) GetBlobIdForString(
	input string,
) (interfaces.BlobId, interfaces.FuncRepool) {
	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	if _, err := io.WriteString(hash, input); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetBlobId()
}

func (hashType HashType) FromStringContent(input string) interfaces.BlobId {
	id, _ := hashType.GetBlobIdForString(input)
	return id
}

func (hashType HashType) FromStringFormat(
	format string,
	args ...any,
) (interfaces.BlobId, interfaces.FuncRepool) {
	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	if _, err := fmt.Fprintf(hash, format, args...); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetBlobId()
}
