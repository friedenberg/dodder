package markl

import (
	"crypto/rand"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	FuncFormatSecGenerate     func(io.Reader) ([]byte, error)
	FuncFormatSecGetPublicKey func(private interfaces.MarklId) ([]byte, error)
	FuncFormatSecGetIOWrapper func(private interfaces.MarklId) (interfaces.IOWrapper, error)
	FuncFormatSecSign         func(sec, mes interfaces.MarklId, readerRand io.Reader) ([]byte, error)

	FormatSec struct {
		id           string
		funcGenerate FuncFormatSecGenerate

		pubkeyFormatId   string
		funcGetPublicKey FuncFormatSecGetPublicKey

		funcGetIOWrapper FuncFormatSecGetIOWrapper

		sigFormatId string
		funcSign    FuncFormatSecSign
	}
)

var _ interfaces.MarklFormat = FormatSec{}

func (format FormatSec) GetMarklFormatId() string {
	return format.id
}

func (format FormatSec) Generate(
	readerRand io.Reader,
	purpose string,
	dst interfaces.MutableMarklId,
) (err error) {
	if format.funcGenerate == nil {
		err = errors.Errorf("format does not support generation: %q", format.id)
		return
	}

	if readerRand == nil {
		readerRand = rand.Reader
	}

	var bites []byte

	if bites, err = format.funcGenerate(readerRand); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetPurpose(purpose); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetMarklId(format.id, bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format FormatSec) GetPublicKey(
	private interfaces.MarklId,
	purpose string,
) (public Id, err error) {
	if format.funcGetPublicKey == nil {
		err = errors.Errorf(
			"format does not support getting public key: %q",
			format.id,
		)
		return
	}

	var bites []byte

	if bites, err = format.funcGetPublicKey(private); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetPurpose(PurposeRepoPubKeyV1); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = public.SetMarklId(format.pubkeyFormatId, bites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format FormatSec) GetIOWrapper(
	private interfaces.MarklId,
) (ioWrapper interfaces.IOWrapper, err error) {
	if format.funcGetIOWrapper == nil {
		err = errors.Errorf(
			"format does not support getting io wrapper key: %q",
			format.id,
		)
		return
	}

	if ioWrapper, err = format.funcGetIOWrapper(private); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func (format FormatSec) Sign(
	sec interfaces.MarklId,
	mes interfaces.MarklId,
	sigDst interfaces.MutableMarklId,
	purpose string,
) (err error) {
	if format.funcSign == nil {
		err = errors.Errorf(
			"format does not support getting signing: %q",
			format.id,
		)
		return
	}

	var sigBites []byte

	if sigBites, err = format.funcSign(sec, mes, nil); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sigDst.SetPurpose(
		purpose,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = sigDst.SetMarklId(format.sigFormatId, sigBites); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
