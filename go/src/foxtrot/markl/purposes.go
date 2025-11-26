package markl

import (
	"fmt"
)

// purposes currently treated as formats
const (
	// TODO move to ids' builtin types
	// and then add registration
	// keep sorted

	// Blob Digests
	PurposeBlobDigestV1 = "dodder-blob-digest-sha256-v1"

	// Object Digests
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
	makePurpose(
		PurposeBlobDigestV1,
		FormatIdHashSha256,
		FormatIdHashBlake2b256,
	)

	makePurpose(
		PurposeObjectDigestV1,
		FormatIdHashSha256,
		FormatIdHashBlake2b256,
	)

	makePurpose(
		PurposeObjectDigestV2,
		FormatIdHashSha256,
		FormatIdHashBlake2b256,
	)

	makePurpose(
		PurposeV5MetadataDigestWithoutTai,
		FormatIdHashSha256,
		FormatIdHashBlake2b256,
	)

	makePurpose(
		PurposeObjectMotherSigV1,
		FormatIdEd25519Sig,
	)

	makePurpose(
		PurposeObjectSigV0,
		FormatIdEd25519Sig,
	)

	makePurpose(
		PurposeObjectSigV1,
		FormatIdEd25519Sig,
	)

	makePurpose(
		PurposeRepoPrivateKeyV1,
		FormatIdEd25519Sec,
	)

	makePurpose(
		PurposeRepoPubKeyV1,
		FormatIdEd25519Pub,
	)

	makePurpose(PurposeRequestAuthChallengeV1)
	makePurpose(PurposeRequestAuthResponseV1)

	makePurpose(
		PurposeMadderPubKeyV1,
		FormatIdEd25519Pub,
	)

	makePurpose(
		PurposeMadderPrivateKeyV0,
		FormatIdEd25519Sec,
		FormatIdAgeX25519Sec,
	)

	makePurpose(
		PurposeMadderPrivateKeyV1,
		FormatIdEd25519Sec,
		FormatIdAgeX25519Sec,
	)
}

var purposes = map[string]Purpose{}

type Purpose struct {
	id        string
	formatIds map[string]struct{}
}

func GetPurpose(purposeId string) Purpose {
	purpose, ok := purposes[purposeId]

	if !ok {
		panic(fmt.Sprintf("no purpose registered for id %q", purposeId))
	}

	return purpose
}

func makePurpose(purposeId string, formatIds ...string) {
	_, alreadyExists := purposes[purposeId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", purposeId))
	}

	purpose := Purpose{
		id:        purposeId,
		formatIds: make(map[string]struct{}),
	}

	for _, formatId := range formatIds {
		_, ok := purpose.formatIds[formatId]

		if ok {
			panic(
				fmt.Sprintf("format id (%q) registered for purpose (%q) more than once",
					formatId,
					purposeId,
				),
			)
		}

		purpose.formatIds[formatId] = struct{}{}
	}

	purposes[purposeId] = purpose
}
