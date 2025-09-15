package markl

import (
	"fmt"
)

// purposes currently treated as formats
const (
	// TODO move to ids' builtin types
	// and then add registration
	// keep sorted

	// Digests
	PurposeObjectDigestV1             = "dodder-object-digest-sha256-v1"
	PurposeObjectDigestV2             = "dodder-object-digest-v2"
	PurposeV5MetadataDigestWithoutTai = "dodder-object-metadata-digest-without_tai-v1"
	// FormatIdObjectDigestObjectId       = "dodder-object-digest-objectId-v1"
	// FormatIdObjectDigestObjectIdTai    =
	// "dodder-object-digest-objectId+tai-v1"

	// Signatures
	PurposeObjectMotherSigV1     = "dodder-object-mother-sig-v1"
	PurposeObjectSigV0           = "dodder-repo-sig-v1"
	PurposeObjectSigV1           = "dodder-object-sig-v1"
	PurposeRequestAuthResponseV1 = "dodder-request_auth-response-v1"
	PurposeRequestRepoSigV1      = "dodder-request_auth-repo-sig-v1"

	// PubKeys
	PurposeRepoPubKeyV1   = "dodder-repo-public_key-v1"
	PurposeMadderPubKeyV1 = "madder-public_key-v1"

	// PrivateKeys
	PurposeRepoPrivateKeyV1   = "dodder-repo-private_key-v1"
	PurposeMadderPrivateKeyV0 = "madder-private_key-v0"
	PurposeMadderPrivateKeyV1 = "madder-private_key-v1"

	// Arbitrary
	PurposeRequestAuthChallengeV1 = "dodder-request_auth-challenge-v1"
)

func init() {
	// purposes that need to be reregistered
	makePurpose(PurposeObjectMotherSigV1)
	makePurpose(PurposeObjectSigV0)
	makePurpose(PurposeObjectSigV1)

	makePurpose(PurposeRepoPrivateKeyV1)
	makePurpose(PurposeRepoPubKeyV1)

	makePurpose(PurposeRequestAuthChallengeV1)
	makePurpose(PurposeRequestAuthResponseV1)

	makePurpose(PurposeMadderPubKeyV1)
	makePurpose(PurposeMadderPrivateKeyV0)
	makePurpose(PurposeMadderPrivateKeyV1)
}

var purposes = map[string]Purpose{}

type Purpose struct {
	id string
}

func makePurpose(purposeId string) {
	_, alreadyExists := purposes[purposeId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", purposeId))
	}

	purposes[purposeId] = Purpose{
		id: purposeId,
	}
}
