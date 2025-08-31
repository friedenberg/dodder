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
	FormatIdObjectDigestSha256V1       = "dodder-object-digest-sha256-v1"
	FormatIdV5MetadataDigestWithoutTai = "dodder-object-metadata-digest-without_tai-v1"

	// Signatures
	FormatIdObjectMotherSigV1     = "dodder-object-mother-sig-v1"
	FormatIdObjectSigV0           = "dodder-repo-sig-v1"
	FormatIdObjectSigV1           = "dodder-object-sig-v1"
	FormatIdRequestAuthResponseV1 = "dodder-request_auth-response-v1"

	// PubKeys
	FormatIdRepoPubKeyV1 = "dodder-repo-public_key-v1"

	// PrivateKeys
	FormatIdRepoPrivateKeyV1 = "dodder-repo-private_key-v1"

	// Arbitrary
	FormatIdRequestAuthChallengeV1 = "dodder-request_auth-challenge-v1"
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

// TODO implement as fields on formats
func SetMerkleIdWithFormat(
	id interfaces.MutableMarklId,
	formatId string,
	data []byte,
) (err error) {
	if err = id.SetFormat(
		formatId,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	var marklTypeId string

	switch formatId {
	case FormatIdRepoPubKeyV1:
		marklTypeId = TypeIdEd25519Pub

	case FormatIdObjectSigV1:
		marklTypeId = TypeIdEd25519Sig

	default:
		err = errors.Errorf("unsupported format: %q", formatId)
		return
	}

	if err = id.SetMerkleId(
		marklTypeId,
		data,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func SetMerkleIdWithFormatBlech32(
	id interfaces.MutableMarklId,
	formatId string,
	blechValue string,
) (err error) {
	if err = id.SetFormat(formatId); err != nil {
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

	marklTypeId := id.GetMarklType()

	switch marklTypeId.GetMarklTypeId() {
	case TypeIdEd25519Sig:
		switch formatId {
		case FormatIdObjectMotherSigV1,
			FormatIdObjectSigV0,
			FormatIdObjectSigV1:
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				formatId,
				blechValue,
			)
			return
		}

	case TypeIdEd25519Pub:
		switch formatId {
		case FormatIdRepoPubKeyV1:
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				formatId,
				blechValue,
			)
			return
		}

	case HashTypeIdSha256:
		switch formatId {
		case FormatIdObjectDigestSha256V1,
			FormatIdV5MetadataDigestWithoutTai,
			"":
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				formatId,
				blechValue,
			)
			return
		}

	case HashTypeIdBlake2b256:
		fallthrough

	default:
		err = errors.Errorf(
			"unsupported format: %q. Value: %q",
			formatId,
			blechValue,
		)
		return
	}

	return
}
