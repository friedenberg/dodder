package merkle

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
	hashTypes      map[string]HashType = map[string]HashType{}
	HashTypeSha256 HashType
)

func init() {
	HashTypeSha256 = makeHashType(
		crypto.SHA256,
		HashTypeIdSha256,
		sha256.New,
		&HashTypeSha256,
	)
}

func makeHashType(
	cryptoHash crypto.Hash,
	tipe string,
	constructor func() hash.Hash,
	self *HashType,
) HashType {
	hashType, alreadyExists := hashTypes[tipe]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", tipe))
	}

	hashType = HashType{
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
	hashType.null.tipe = tipe
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

var _ interfaces.HashType = HashType{}

func (hashType *HashType) Get() *Hash {
	hash := hashType.pool.Get()
	hash.hashType = hashType
	return &hash
}

func (hashType HashType) Put(hash *Hash) {
	errors.PanicIfError(MakeErrWrongType(hashType.tipe, hash.GetType()))
	hashType.pool.Put(*hash)
}

func (hashType HashType) GetType() string {
	return hashType.tipe
}

func (hashType HashType) GetSize() int {
	return hashType.Hash.Size()
}

func (hashType HashType) GetBlobId() (interfaces.MutableBlobId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

	return hash.GetBlobId()
}

func (hashType HashType) GetBlobIdForString(
	input string,
) (interfaces.BlobId, interfaces.FuncRepool) {
	hash := hashType.Get()
	defer hashType.Put(hash)

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
	hash := hashType.Get()
	defer hashType.Put(hash)

	if _, err := fmt.Fprintf(hash, format, args...); err != nil {
		errors.PanicIfError(err)
	}

	return hash.GetBlobId()
}

func (hashType HashType) GetBlobIdForHexString(
	input string,
) (interfaces.BlobId, interfaces.FuncRepool) {
	hash := hashType.pool.Get()
	defer hashType.pool.Put(hash)

	id, repool := hash.GetBlobId()

	errors.PanicIfError(SetHexBytes(hashType.tipe, id, []byte(input)))

	return id, repool
}
