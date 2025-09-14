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
	tipe := id.GetMarklFormat()
	data := id.GetBytes()

	if tipe == nil && len(data) == 0 {
		return ""
	}

	if len(data) == 0 {
		return ""
	} else {
		typeId := tipe.GetMarklFormatId()
		combined := make([]byte, len(typeId)+len(data))
		copy(combined, []byte(typeId))
		copy(combined[len(typeId):], data)
		bites, err := blech32.EncodeDataOnly(combined)
		errors.PanicIfError(err)
		return string(bites)
	}
}

func SetBlechCombinedHRPAndData(
	id interfaces.MutableMarklId,
	value string,
) (err error) {
	var typeId string

	var typeIdAndData []byte

	if typeIdAndData, err = blech32.DecodeDataOnly([]byte(value)); err != nil {
		err = errors.Wrapf(err, "Value: %q", value)
		return
	}

	if bytes.HasPrefix(typeIdAndData, []byte(HashTypeIdSha256)) {
		typeId = HashTypeIdSha256
	} else if bytes.HasPrefix(typeIdAndData, []byte(HashTypeIdBlake2b256)) {
		typeId = HashTypeIdBlake2b256
	} else {
		err = errors.Errorf("unsupported format: %x", typeIdAndData)
		return
	}

	data := typeIdAndData[len(typeId):]

	if err = id.SetMarklId(typeId, data); err != nil {
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
		HashTypeIdSha256,
		decodedBytes,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
