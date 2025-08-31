package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

const (
	// TODO move to ids' builtin types
	// and then add registration
	// keep sorted
	FormatIdObjectDigestSha256V1       = "dodder-object-digest-sha256-v1"
	FormatIdObjectMotherSigV1          = "dodder-object-mother-sig-v1"
	FormatIdObjectSigV0                = "dodder-repo-sig-v1"
	FormatIdObjectSigV1                = "dodder-object-sig-v1"
	FormatIdRepoPrivateKeyV1           = "dodder-repo-private_key-v1"
	FormatIdRepoPubKeyV1               = "dodder-repo-public_key-v1"
	FormatIdRequestAuthChallengeV1     = "dodder-request_auth-challenge-v1"
	FormatIdRequestAuthResponseV1      = "dodder-request_auth-response-v1"
	FormatIdV5MetadataDigestWithoutTai = "dodder-object-metadata-digest-without_tai-v1"
	TypeIdEd25519                      = "ed25519"
)

func GetMarklTypeOrError(typeId string) (interfaces.MarklType, error) {
	tipe, ok := types[typeId]

	if !ok {
		err := errors.Errorf("unknown type: %q", typeId)
		return nil, err
	}

	return tipe, nil
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

type fakeHashType struct {
	typeId string
}

var _ interfaces.MarklType = fakeHashType{}

func (tipe fakeHashType) GetMarklTypeId() string {
	return tipe.typeId
}

func makeFakeHashType(typeId string) {
	_, alreadyExists := types[typeId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", typeId))
	}

	tipe := fakeHashType{
		typeId: typeId,
	}

	types[typeId] = tipe
}
