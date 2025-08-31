package markl

import (
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

func (id Id) MarshalText() (bites []byte, err error) {
	if id.tipe.GetMarklTypeId() == HashTypeIdSha256 {
		bites = fmt.Appendf(nil, "%x", id.data)
	} else {
		if bites, err = blech32.Encode(id.tipe.GetMarklTypeId(), id.data); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	return
}

func (id *Id) UnmarshalText(bites []byte) (err error) {
	var typeId string
	if typeId, id.data, err = blech32.Decode(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	if id.tipe, err = GetMarklTypeOrError(typeId); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
