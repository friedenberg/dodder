package markl

import (
	"crypto/sha256"
	"fmt"
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"golang.org/x/crypto/blake2b"
)

const (
	HashTypeIdSha256     = "sha256"
	HashTypeIdBlake2b256 = "blake2b256"
)

var (
	hashTypes map[string]HashType = map[string]HashType{}

	// TODO remove unnecessary references
	HashTypeSha256     HashType
	HashTypeBlake2b256 HashType
)

func init() {
	HashTypeSha256 = makeHashType(
		sha256.New,
		HashTypeIdSha256,
		&HashTypeSha256,
	)

	HashTypeBlake2b256 = makeHashType(
		func() hash.Hash {
			hash, _ := blake2b.New256(nil)
			return hash
		},
		HashTypeIdBlake2b256,
		&HashTypeBlake2b256,
	)
}

func makeHashType(
	constructor func() hash.Hash,
	id string,
	self *HashType,
) HashType {
	_, alreadyExists := formats[id]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", id))
	}

	hashType := HashType{
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
		id: id,
	}

	hash := constructor()
	hashType.null.format = self
	hashType.null.allocDataIfNecessary(hash.Size())
	hashType.null.data = hash.Sum(hashType.null.data)

	formats[id] = hashType
	hashTypes[id] = hashType

	return hashType
}

func GetHashTypeOrError(typeId string) (hashType HashType, err error) {
	var ok bool
	hashType, ok = hashTypes[typeId]

	if !ok {
		err = errors.Errorf("unknown type: %q", typeId)
		return
	}

	return
}
