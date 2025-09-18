package markl

import (
	"fmt"
	"strings"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

func (id Id) MarshalText() (bites []byte, err error) {
	if id.format == nil {
		return
	}

	var hrp string

	if prupose := id.GetPurpose(); prupose != "" {
		hrp = fmt.Sprintf(
			"%s@%s",
			id.GetPurpose(),
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

func (id *Id) UnmarshalText(bites []byte) (err error) {
	if len(bites) == 0 {
		id.Reset()
		return
	}

	var purposeAndFormatId string

	if purposeAndFormatId, id.data, err = blech32.Decode(bites); err != nil {
		err = errors.Wrapf(err, "Raw: %q", string(bites))
		return
	}

	purpose, formatId, ok := strings.Cut(purposeAndFormatId, "@")

	if ok {
		if err = id.SetPurpose(purpose); err != nil {
			err = errors.Wrap(err)
			return
		}
	} else {
		formatId = purposeAndFormatId
	}

	if id.format, err = GetFormatOrError(formatId); err != nil {
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
