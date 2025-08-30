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

var hrpToHashType = map[string]*HashType{
	HRPObjectBlobDigestSha256V0: &HashTypeSha256,
	HRPObjectBlobDigestSha256V1: &HashTypeSha256,
	HRPObjectDigestSha256V1:     &HashTypeSha256,
}

func GetHashTypeOrError(tipe string) (*HashType, error) {
	hashType, ok := hrpToHashType[tipe]

	if !ok {
		err := errors.Errorf("unknown type: %q", tipe)
		return nil, err
	}

	return hashType, nil
}

type fakeHashType struct {
	tipe string
}

var _ interfaces.MarklType = fakeHashType{}

func (tipe fakeHashType) GetType() string {
	return tipe.tipe
}

func makeFakeHashType(tipe string) {
	_, alreadyExists := hashTypes[tipe]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", tipe))
	}

	hashType := fakeHashType{
		tipe: tipe,
	}

	hashTypes[tipe] = hashType
}
