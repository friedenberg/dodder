package markl

import (
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/charlie/ohio"
)

func (id *Id) GeneratePrivateKey(
	readerRand io.Reader,
	formatId string,
	purpose string,
) (err error) {
	if formatId != "" {
		if err = id.setFormatId(formatId); err != nil {
			err = errors.Wrap(err)
			return err
		}
	}

	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Wrap(ErrFormatOperationNotSupported{
				Format:        id.format,
				FormatId:      formatId,
				OperationName: "GeneratePrivateKey",
			})

			return err
		}
	}

	if formatSec.Generate == nil {
		err = errors.Wrap(ErrFormatOperationNotSupported{
			Format:        id.format,
			OperationName: "GeneratePrivateKey",
		})

		return err
	}

	if readerRand == nil {
		readerRand = rand.Reader
	}

	var bites []byte

	if bites, err = formatSec.Generate(readerRand); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = id.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = id.SetMarklId(formatSec.Id, bites); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}

func (id Id) GetPublicKey(
	purpose string,
) (public Id, err error) {
	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Wrap(ErrFormatOperationNotSupported{
				Format:        id.format,
				OperationName: "GetPublicKey",
			})

			return public, err
		}
	}

	if formatSec.GetPublicKey == nil {
		err = errors.Wrap(ErrFormatOperationNotSupported{
			Format:        id.format,
			OperationName: "GetPublicKey",
		})

		return public, err
	}

	var bites []byte

	if bites, err = formatSec.GetPublicKey(id); err != nil {
		err = errors.Wrap(err)
		return public, err
	}

	if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
		err = errors.Wrap(err)
		return public, err
	}

	if err = public.SetMarklId(formatSec.PubFormatId, bites); err != nil {
		err = errors.Wrap(err)
		return public, err
	}

	return public, err
}

func (id Id) GetIOWrapper() (ioWrapper interfaces.IOWrapper, err error) {
	if id.IsNull() {
		ioWrapper = ohio.NopeIOWrapper{}
		return ioWrapper, err
	}

	var formatSec FormatSec

	{
		var ok bool

		if formatSec, ok = id.format.(FormatSec); !ok {
			err = errors.Wrap(ErrFormatOperationNotSupported{
				Format:        id.format,
				OperationName: "GetIOWrapper",
			})

			return ioWrapper, err
		}
	}

	if formatSec.GetIOWrapper == nil {
		err = errors.Wrap(ErrFormatOperationNotSupported{
			Format:        id.format,
			OperationName: "GetIOWrapper",
		})

		return ioWrapper, err
	}

	if ioWrapper, err = formatSec.GetIOWrapper(id); err != nil {
		err = errors.Wrap(err)
		return ioWrapper, err
	}

	return ioWrapper, err
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
			err = errors.Wrap(ErrFormatOperationNotSupported{
				Format:        id.format,
				OperationName: "Sign",
			})

			return err
		}
	}

	if formatSec.Sign == nil {
		err = errors.Wrap(ErrFormatOperationNotSupported{
			Format:        id.format,
			OperationName: "Sign",
		})

		return err
	}

	var sigBites []byte

	if sigBites, err = formatSec.Sign(id, mes, nil); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = sigDst.SetPurpose(
		sigPurpose,
	); err != nil {
		err = errors.Wrap(err)
		return err
	}

	if err = sigDst.SetMarklId(formatSec.SigFormatId, sigBites); err != nil {
		err = errors.Wrap(err)
		return err
	}

	return err
}
