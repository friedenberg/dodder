package markl

import (
	"bytes"
	"encoding/hex"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

func StringHRPCombined(id interfaces.MarklId) string {
	format := id.GetMarklFormat()
	data := id.GetBytes()

	if format == nil && len(data) == 0 {
		return ""
	}

	if len(data) == 0 {
		return ""
	} else {
		formatId := format.GetMarklFormatId()
		combined := make([]byte, len(formatId)+len(data))
		copy(combined, []byte(formatId))
		copy(combined[len(formatId):], data)
		bites, err := blech32.EncodeDataOnly(combined)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func SetBlechCombinedHRPAndData(
	id interfaces.MutableMarklId,
	value string,
) (err error) {
	var formatId string

	var formatIdAndData []byte

	if formatIdAndData, err = blech32.DecodeDataOnly([]byte(value)); err != nil {
		err = errors.Wrapf(err, "Value: %q", value)
		return
	}

	if bytes.HasPrefix(formatIdAndData, []byte(FormatIdHashSha256)) {
		formatId = FormatIdHashSha256
	} else if bytes.HasPrefix(formatIdAndData, []byte(FormatIdHashBlake2b256)) {
		formatId = FormatIdHashBlake2b256
	} else {
		err = errors.Errorf("unsupported format: %x", formatIdAndData)
		return
	}

	data := formatIdAndData[len(formatId):]

	if err = id.SetMarklId(formatId, data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

// TODO remove
func SetMaybeSha256(id interfaces.MutableMarklId, value string) (err error) {
	switch len(value) {
	case 65:
		if value[0] != '@' {
			err = errors.Errorf("unknown format: %q", value)
			return
		}

		value = strings.TrimPrefix(value, "@")
		fallthrough

	case 64:
		if err = SetSha256(id, value); err != nil {
			err = errors.Wrap(err)
			return
		}

	default:
		if err = id.Set(value); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

// TODO remove
func SetSha256(id interfaces.MutableMarklId, value string) (err error) {
	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "%q", value)
		return
	}

	if err = id.SetMarklId(
		FormatIdHashSha256,
		decodedBytes,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
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
	case FormatIdSigEd25519:
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

	case FormatIdPubEd25519:
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

	case FormatIdHashSha256:
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

	case FormatIdHashBlake2b256:
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
