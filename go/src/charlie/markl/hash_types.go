package markl

import (
	"crypto/sha256"
	"fmt"
	"hash"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/pool"
	"golang.org/x/crypto/blake2b"
)

const (
	HashTypeIdSha256     = "sha256"
	HashTypeIdBlake2b256 = "blake2b256"
)

var (
	types     map[string]interfaces.MarklType = map[string]interfaces.MarklType{}
	hashTypes map[string]HashType             = map[string]HashType{}

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

	makeType(FormatIdObjectMotherSigV1)
	makeType(FormatIdObjectSigV0)
	makeType(FormatIdObjectSigV1)
	makeType(FormatIdRepoPrivateKeyV1)
	makeType(FormatIdRepoPubKeyV1)
	makeType(FormatIdRequestAuthChallengeV1)
	makeType(FormatIdRequestAuthResponseV1)

	makeType(TypeIdEd25519)
}

func makeHashType(
	constructor func() hash.Hash,
	tipe string,
	self *HashType,
) HashType {
	_, alreadyExists := types[tipe]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", tipe))
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
		tipe: tipe,
	}

	hash := constructor()
	hashType.null.tipe = self
	hashType.null.allocDataIfNecessary(hash.Size())
	hashType.null.data = hash.Sum(hashType.null.data)

	types[tipe] = hashType
	hashTypes[tipe] = hashType

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
