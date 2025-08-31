package markl

import (
	"fmt"

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
)

func init() {
	makeType(FormatIdObjectMotherSigV1)
	makeType(FormatIdObjectSigV0)
	makeType(FormatIdObjectSigV1)
	makeType(FormatIdRepoPrivateKeyV1)
	makeType(FormatIdRepoPubKeyV1)
	makeType(FormatIdRequestAuthChallengeV1)
	makeType(FormatIdRequestAuthResponseV1)
}

type tipe struct {
	typeId string
}

var _ interfaces.MarklType = tipe{}

func (tipe tipe) GetMarklTypeId() string {
	return tipe.typeId
}

func makeType(typeId string) {
	_, alreadyExists := types[typeId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", typeId))
	}

	types[typeId] = tipe{
		typeId: typeId,
	}
}
