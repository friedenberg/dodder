package markl

import (
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

type PrivateKeyGenerator interface {
	GeneratePrivateKey() (err error)
}

func (id *Id) GeneratePrivateKey(
	readerRand io.Reader,
	formatId string,
	purpose string,
) (err error) {
	if formatId != "" {
		if err = id.setFormatId(formatId); err != nil {
			err = errors.Wrap(err)
			return
		}
	}

	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Errorf(
				"id format does not support sec operation: %T",
				id.format,
			)
			return
		}
	}

	if formatSec.funcGenerate == nil {
		err = errors.Errorf(
			"format does not support generation: %q",
			formatSec.id,
		)
		return
	}

	if readerRand == nil {
		readerRand = rand.Reader
	}

	var bites []byte

	if bites, err = formatSec.funcGenerate(readerRand); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMarklId(formatSec.id, bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id Id) GetPublicKey(
	purpose string,
) (public Id, err error) {
	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Errorf(
				"id format does not support sec operation: %T",
				id.format,
			)
			return
		}
	}

	if formatSec.funcGetPublicKey == nil {
		err = errors.Errorf(
			"format does not support getting public key: %q",
			formatSec.id,
		)
		return
	}

	var bites []byte

	if bites, err = formatSec.funcGetPublicKey(id); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetMarklId(formatSec.pubkeyFormatId, bites); err != nil {
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
