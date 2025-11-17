package markl

import (
	"bytes"
	"encoding/hex"
	"strings"

	"code.linenisgreat.com/dodder/go/src/_/interfaces"
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
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
		return err
	}

	if bytes.HasPrefix(formatIdAndData, []byte(FormatIdHashSha256)) {
		formatId = FormatIdHashSha256
	} else if bytes.HasPrefix(formatIdAndData, []byte(FormatIdHashBlake2b256)) {
		formatId = FormatIdHashBlake2b256
	} else {
		err = errors.Errorf("unsupported format: %x", formatIdAndData)
		return err
	}

	data := formatIdAndData[len(formatId):]

	if err = id.SetMarklId(formatId, data); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO remove
func SetMaybeSha256(id interfaces.MutableMarklId, value string) (err error) {
	// TODO use registered format lengths
	switch len(value) {
	case 65:
		if value[0] != '@' {
			err = errors.Errorf("unknown format: %q", value)
			return err
		}

		value = strings.TrimPrefix(value, "@")
		fallthrough

	case 64:
		if err = setSha256(id, value); err != nil {
			err = errors.Wrap(err)
			return err
		}

	default:
		if err = id.Set(value); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	return err
}

// TODO remove
func setSha256(id interfaces.MutableMarklId, value string) (err error) {
	var decodedBytes []byte

	if decodedBytes, err = hex.DecodeString(value); err != nil {
		err = errors.Wrapf(err, "%q", value)
		return err
	}

	if err = id.SetMarklId(
		FormatIdHashSha256,
		decodedBytes,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

// TODO use type and format registrations
// TODO need to set format
func SetMarklIdWithFormatBlech32(
	id interfaces.MutableMarklId,
	purpose string,
	blechValue string,
) (err error) {
	if err = id.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = id.Set(
		blechValue,
	); err != nil {
		if errors.Is(err, blech32.ErrSeparatorMissing) {
			if err = setSha256(
				id,
				blechValue,
			); err != nil {
				err = errors.Wrap(err)
				return err
			}
		} else {
			err = errors.Wrap(err)
			return err
		}
	}

	marklTypeId := id.GetMarklFormat()

	switch marklTypeId.GetMarklFormatId() {
	case FormatIdEd25519Sig:
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
			return err
		}

	case FormatIdEd25519Pub:
		switch purpose {
		case PurposeRepoPubKeyV1:
			break

		default:
			err = errors.Errorf(
				"unsupported format: %q. Value: %q",
				purpose,
				blechValue,
			)
			return err
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
			return err
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
			return err
		}

	default:
		err = errors.Errorf(
			"unsupported format: %q. Value: %q",
			purpose,
			blechValue,
		)
		return err
	}

	return err
}
