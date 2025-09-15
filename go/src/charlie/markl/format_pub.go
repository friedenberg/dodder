package markl

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type (
	FuncFormatPubVerify func(pubkey, message, sig interfaces.MarklId) error

	FormatPub struct {
		id string

		funcVerify FuncFormatPubVerify
	}
)

var _ interfaces.MarklFormat = FormatPub{}

func (format FormatPub) GetMarklFormatId() string {
	return format.id
}

func (format FormatPub) Verify(
	publicKey, message, sig interfaces.MarklId,
) (err error) {
	if format.funcVerify == nil {
		err = errors.Errorf(
			"format does not support verification: %q",
			format.id,
		)
		return
	}

	defer errors.DeferredRecover(&err)

	if err = format.funcVerify(publicKey, message, sig); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}
