package markl

import (
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
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

type PrivateKeyGenerator interface {
	GeneratePrivateKey() (err error)
}

func GetPublicKey(
	private interfaces.MarklId,
	purpose string,
) (public Id, err error) {
	var formatSec FormatSec

	if formatSec, err = GetFormatSecOrError(private); err != nil {
		err = errors.Wrap(err)
		return
	}

	if public, err = formatSec.GetPublicKey(
		private,
		purpose,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func GetIOWrapper(
	sec interfaces.MarklId,
) (ioWrapper interfaces.IOWrapper, err error) {
	if sec.IsNull() {
		ioWrapper = files.NopeIOWrapper{}
		return
	}

	var formatSec FormatSec

	if formatSec, err = GetFormatSecOrError(sec); err != nil {
		err = errors.Wrap(err)
		return
	}

	if ioWrapper, err = formatSec.GetIOWrapper(
		sec,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func Sign(
	sec interfaces.MarklId,
	mes interfaces.MarklId,
	sigDst interfaces.MutableMarklId,
	sigPurpose string,
) (err error) {
	var formatSec FormatSec

	if formatSec, err = GetFormatSecOrError(sec); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = formatSec.Sign(sec, mes, sigDst, sigPurpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
