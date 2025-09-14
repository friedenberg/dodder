package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

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
	makeType(PurposeObjectMotherSigV1)
	makeType(PurposeObjectSigV0)
	makeType(PurposeObjectSigV1)

	makeType(PurposeRepoPrivateKeyV1)
	makeType(PurposeRepoPubKeyV1)

	makeType(PurposeRequestAuthChallengeV1)
	makeType(PurposeRequestAuthResponseV1)

	makeType(PurposeMadderPubKeyV1)
	makeType(PurposeMadderPrivateKeyV1)
}

type format struct {
	id string
}

var _ interfaces.MarklFormat = format{}

func (format format) GetMarklFormatId() string {
	return format.id
}

func makeType(formatId string) {
	_, alreadyExists := types[formatId]

	if alreadyExists {
		panic(fmt.Sprintf("hash type already registered: %q", formatId))
	}

	types[formatId] = format{
		id: formatId,
	}
}

// TODO use type and format registrations
func SetMarklIdWithFormatBlech32(
	id interfaces.MutableMarklId,
	purpose string,
	blechValue string,
) (err error) {
	if err = id.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.Set(
		blechValue,
	); err != nil {
		if errors.Is(err, blech32.ErrSeparatorMissing) {
			if err = SetSha256(
				id,
				blechValue,
			); err != nil {
				err = errors.Wrap(err)
				return
			}
		} else {
			err = errors.Wrap(err)
			return
		}
	}

	marklTypeId := id.GetMarklFormat()

	switch marklTypeId.GetMarklFormatId() {
	case TypeIdEd25519Sig:
		switch purpose {
		case PurposeObjectMotherSigV1,
			PurposeObjectSigV0,
			PurposeObjectSigV1:
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				purpose,
				blechValue,
			)
			return
		}

	case TypeIdEd25519Pub:
		switch purpose {
		case PurposeRepoPubKeyV1:
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				purpose,
				blechValue,
			)
			return
		}

	case HashTypeIdSha256:
		switch purpose {
		case PurposeObjectDigestV1,
			PurposeV5MetadataDigestWithoutTai,
			"":
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				purpose,
				blechValue,
			)
			return
		}

	case HashTypeIdBlake2b256:
		switch purpose {
		case PurposeObjectDigestV1,
			PurposeV5MetadataDigestWithoutTai,
			"":
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				purpose,
				blechValue,
			)
			return
		}

	default:
		err = errors.Errorf(
			"unsupported format: %q. Value: %q",
			purpose,
			blechValue,
		)
		return
	}

	return
}
