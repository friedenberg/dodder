package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

func GeneratePrivateKey(
	rand io.Reader,
	purpose string,
	formatId string,
	dst interfaces.MutableMarklId,
) (err error) {
	var formatSec FormatSec

	if formatSec, err = GetFormatSecOrError(FormatId(formatId)); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = formatSec.Generate(
		rand,
		purpose,
		dst,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
