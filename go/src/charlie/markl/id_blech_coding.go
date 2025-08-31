package markl

import (
	"bytes"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

func StringHRPCombined(id interfaces.MarklId) string {
	tipe := id.GetMarklType()
	data := id.GetBytes()

	if tipe == nil && len(data) == 0 {
		return ""
	}

	if len(data) == 0 {
		return ""
	} else {
		typeId := tipe.GetMarklTypeId()
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

	if err = id.SetMerkleId(typeId, data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
