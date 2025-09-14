package markl

import (
	"fmt"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type HashType struct {
	pool interfaces.PoolValue[Hash]
	formatId string
	null Id
}

var (
	_ interfaces.MarklFormat = HashType{}
	_ interfaces.HashType  = HashType{}
)

func (hashType HashType) GetHash() interfaces.Hash {
	return hashType.Get()
}

func (hashType HashType) PutHash(hash interfaces.Hash) {
	if correctHashType, ok := hash.(*Hash); ok {
		hashType.Put(correctHashType)
	} else {
		panic(errors.Errorf("expected type %T but got %T", correctHashType, hash))
	}
}

func (hashType *HashType) Get() *Hash {
	hash := hashType.pool.Get()
	hash.hashType = hashType
	return &hash
}

func (hashType HashType) Put(hash *Hash) {
	errors.PanicIfError(
		MakeErrWrongType(hashType.formatId, hash.GetMarklFormat().GetMarklFormatId()),
	)
	hashType.pool.Put(*hash)
}

func (hashType HashType) GetMarklFormatId() string {
	return hashType.formatId
}

func (hashType HashType) GetSize() int {
	return hashType.null.GetSize()
}

func (hashType HashType) GetBlobId() (interfaces.MutableMarklId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

	return hash.GetMarklId()
}

func (hashType HashType) GetMarklIdForString(
	input string,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

	if _, err := io.WriteString(hash, input); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (hashType HashType) GetMarklIdForMarklId(
	input interfaces.MarklId,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

	if _, err := hash.Write(input.GetBytes()); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (hashType HashType) FromStringContent(input string) interfaces.MarklId {
	id, _ := hashType.GetMarklIdForString(input)
	return id
}

func (hashType HashType) GetMarklIdFromStringFormat(
	format string,
	args ...any,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

	if _, err := fmt.Fprintf(hash, format, args...); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetMarklId()
}

func (hashType HashType) GetBlobIdForHexString(
	input string,
) (interfaces.MarklId, interfaces.FuncRepool) {
	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	id, repool := hash.GetMarklId()

	errors.PanicIfError(SetHexBytes(hashType.formatId, id, []byte(input)))

	return id, repool
}
