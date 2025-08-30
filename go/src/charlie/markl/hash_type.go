package markl

import (
	"crypto"
	"crypto/sha256"
	"fmt"
	"hash"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
)

const (
	HashTypeIdSha256 = "sha256"
)

var (
	hashTypes      map[string]interfaces.MarklType = map[string]interfaces.MarklType{}
	HashTypeSha256 HashType
)

func init() {
	HashTypeSha256 = makeHashType(
		crypto.SHA256,
		HashTypeIdSha256,
		sha256.New,
		&HashTypeSha256,
	)

	makeFakeHashType(HRPObjectBlobDigestSha256V1)
	makeFakeHashType(HRPObjectDigestSha256V1)
	makeFakeHashType(HRPObjectMotherSigV1)
	makeFakeHashType(HRPObjectSigV0)
	makeFakeHashType(HRPObjectSigV1)
	makeFakeHashType(HRPRepoPrivateKeyV1)
	makeFakeHashType(HRPRepoPubKeyV1)
	makeFakeHashType(HRPRequestAuthChallengeV1)
	makeFakeHashType(HRPRequestAuthResponseV1)
}

func makeHashType(
	cryptoHash crypto.Hash,
	tipe string,
	constructor func() hash.Hash,
	self *HashType,
) HashType {
	_, alreadyExists := hashTypes[tipe]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", tipe))
	}

	hashType := HashType{
		Hash: cryptoHash,
		pool: pool.MakeValue(
			func() Hash {
				return Hash{
					hash:     constructor(),
					hashType: self,
				}
			},
			func(hash Hash) {
				hash.Reset()
			},
		),
		tipe: tipe,
	}

	hash := constructor()
	hashType.null.tipe = self
	hashType.null.allocDataIfNecessary(cryptoHash.Size())
	hashType.null.data = hash.Sum(hashType.null.data)

	hashTypes[tipe] = hashType

	return hashType
}

type HashType struct {
	crypto.Hash
	pool interfaces.PoolValue[Hash]
	tipe string
	null Id
}

var _ interfaces.MarklType = HashType{}

func (hashType *HashType) Get() *Hash {
	hash := hashType.pool.Get()
	hash.hashType = hashType
	return &hash
}

func (hashType HashType) Put(hash *Hash) {
	errors.PanicIfError(
		MakeErrWrongType(hashType.tipe, hash.GetType().GetType()),
	)
	hashType.pool.Put(*hash)
}

func (hashType HashType) GetType() string {
	return hashType.tipe
}

func (hashType HashType) GetSize() int {
	return hashType.Hash.Size()
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

func (hashType HashType) FromStringContent(input string) interfaces.MarklId {
	id, _ := hashType.GetMarklIdForString(input)
	return id
}

func (hashType HashType) FromStringFormat(
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

	errors.PanicIfError(SetHexBytes(hashType.tipe, id, []byte(input)))

	return id, repool
}
