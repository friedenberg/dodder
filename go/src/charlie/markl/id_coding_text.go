package markl

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

func (id Id) MarshalText() (bites []byte, err error) {
	if id.tipe == nil {
		return
	}

	var hrp string

	if format := id.GetFormat(); format != "" {
		hrp = fmt.Sprintf(
			"%s@%s",
			id.GetFormat(),
			id.tipe.GetMarklTypeId(),
		)
	} else {
		hrp = id.tipe.GetMarklTypeId()
	}

	if bites, err = blech32.Encode(hrp, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *Id) UnmarshalText(bites []byte) (err error) {
	if len(bites) == 0 {
		id.Reset()
		return
	}

	var formatAndTypeId string

	if formatAndTypeId, id.data, err = blech32.Decode(bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	formatId, typeId, ok := strings.Cut(formatAndTypeId, "@")

	if ok {
		if err = id.SetFormat(formatId); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		typeId = formatAndTypeId
	}

	if id.tipe, err = GetMarklTypeOrError(typeId); err != nil {
		err = errors.Wrapf(
			err,
			"failed to unmarshal %T with contents: %s",
			id,
			bites,
		)
		return
	}

	return
}
