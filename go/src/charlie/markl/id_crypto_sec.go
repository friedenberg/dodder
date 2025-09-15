package markl

import (
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/files"
)

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

	if formatSec.Generate == nil {
		err = errors.Errorf(
			"format does not support generation: %q",
			formatSec.Id,
		)
		return
	}

	if readerRand == nil {
		readerRand = rand.Reader
	}

	var bites []byte

	if bites, err = formatSec.Generate(readerRand); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = id.SetMarklId(formatSec.Id, bites); err != nil {
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

	if formatSec.GetPublicKey == nil {
		err = errors.Errorf(
			"format does not support getting public key: %q",
			formatSec.Id,
		)
		return
	}

	var bites []byte

	if bites, err = formatSec.GetPublicKey(id); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetMarklId(formatSec.PubFormatId, bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id Id) GetIOWrapper() (ioWrapper interfaces.IOWrapper, err error) {
	if id.IsNull() {
		ioWrapper = files.NopeIOWrapper{}
		return
	}

	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Errorf(
				"id format does not support sec operations: %T",
				id.format,
			)
			return
		}
	}

	if formatSec.GetIOWrapper == nil {
		err = errors.Errorf(
			"format does not support getting io wrapper key: %q",
			formatSec.Id,
		)
		return
	}

	if ioWrapper, err = formatSec.GetIOWrapper(id); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (id Id) Sign(
	mes interfaces.MarklId,
	sigDst interfaces.MutableMarklId,
	sigPurpose string,
) (err error) {
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

	if formatSec.Sign == nil {
		err = errors.Errorf(
			"format does not support signing: %q",
			formatSec.Id,
		)
		return
	}

	var sigBites []byte

	if sigBites, err = formatSec.Sign(id, mes, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sigDst.SetPurpose(
		sigPurpose,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sigDst.SetMarklId(formatSec.SigFormatId, sigBites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
