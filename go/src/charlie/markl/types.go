package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

const (
	// keep sorted
	// TODO move to ids' builtin types
	// and then add registration
	HRPObjectBlobDigestSha256V0 = HashTypeIdSha256
	HRPObjectBlobDigestSha256V1 = "dodder-object-blob-digest-sha256-v1"
	HRPObjectDigestSha256V1     = "dodder-object-digest-sha256-v1"
	HRPObjectMotherSigV1        = "dodder-object-mother-sig-v1"
	HRPObjectSigV0              = "dodder-repo-sig-v1"
	HRPObjectSigV1              = "dodder-object-sig-v1"
	HRPRepoPrivateKeyV1         = "dodder-repo-private_key-v1"
	HRPRepoPubKeyV1             = "dodder-repo-public_key-v1"
	HRPRequestAuthChallengeV1   = "dodder-request_auth-challenge-v1"
	HRPRequestAuthResponseV1    = "dodder-request_auth-response-v1"
)

var hrpValid = []string{
	// keep sorted
	HRPObjectBlobDigestSha256V0,
	HRPObjectBlobDigestSha256V1,
	HRPObjectDigestSha256V1,
	HRPObjectMotherSigV1,
	HRPObjectSigV0,
	HRPObjectSigV1,
	HRPRepoPrivateKeyV1,
	HRPRepoPubKeyV1,
	HRPRequestAuthChallengeV1,
	HRPRequestAuthResponseV1,
}

var typeLookup = map[string]*HashType{
	HRPObjectBlobDigestSha256V0: &HashTypeSha256,
	HRPObjectBlobDigestSha256V1: &HashTypeSha256,
	HRPObjectDigestSha256V1:     &HashTypeSha256,
}

func GetMarklTypeOrError(typeId string) (interfaces.MarklType, error) {
	tipe, ok := types[typeId]

	if !ok {
		err := errors.Errorf("unknown type: %q", typeId)
		return nil, err
	}

	return tipe, nil
}

func GetHashTypeOrError(typeId string) (hashType HashType, err error) {
	tipe, ok := types[typeId]

	if !ok {
		err = errors.Errorf("unknown type: %q", typeId)
		return
	}

	hashType = tipe.(HashType)

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
