package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func Verify(
	pub, mes, sig interfaces.MarklId,
) (err error) {
	var formatPub FormatPub

	if formatPub, err = GetFormatPubOrError(pub); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = formatPub.Verify(pub, mes, sig); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
