package markl

import (
	"bytes"
	"fmt"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
	"code.linenisgreat.com/zit/go/zit/src/bravo/bech32"
)

type IdBroken Id

func (id IdBroken) MarshalText() (bites []byte, err error) {
	if id.format == nil {
		return
	}

	var hrp string

	if prupose := Id(id).GetPurpose(); prupose != "" {
		hrp = fmt.Sprintf(
			"%s@%s",
			Id(id).GetPurpose(),
			id.format.GetMarklFormatId(),
		)
	} else {
		hrp = id.format.GetMarklFormatId()
	}

	if bites, err = blech32.Encode(hrp, id.data); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id *IdBroken) UnmarshalText(bites []byte) (err error) {
	if len(bites) == 0 {
		((*Id)(id)).Reset()
		return
	}

	purpose, _, ok := bytes.Cut(bites, []byte{'@'})

	var hrp string

	if hrp, id.data, err = bech32.Decode(string(bites)); err != nil {
		err = errors.Wrapf(err, "Raw: %q", string(bites))
		return
	}

	ui.Debug().Print(ok, purpose, hrp)

	// if ok {
	// 	if err = ((*Id)(id)).SetPurpose(purpose); err != nil {
	// 		err = errors.Wrap(err)
	// 		return
	// 	}
	// } else {
	// 	formatId = hrp
	// }

	// if id.format, err = GetFormatOrError(formatId); err != nil {
	// 	err = errors.Wrapf(
	// 		err,
	// 		"failed to unmarshal %T with contents: %s",
	// 		id,
	// 		bites,
	// 	)
	// 	return
	// }

	return
}
