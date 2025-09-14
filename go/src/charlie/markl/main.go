package markl

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
)

type PrivateKeyGenerator interface {
	GeneratePrivateKey() (err error)
}

func Sign(
	key interfaces.MarklId,
	message interfaces.MarklId,
	format string,
	tipe string,
	dst interfaces.MutableMarklId,
) (err error) {
	switch key.GetMarklFormat().GetMarklFormatId() {
	default:
		err = errors.Errorf("not a private key: %q", key.StringWithFormat())
		return

	case TypeIdEd25519Sec:
	}

	privateKey := ed25519.PrivateKey(key.GetBytes())

	var sig []byte

	if sig, err = privateKey.Sign(
		nil,
		message.GetBytes(),
		&ed25519.Options{},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetPurpose(
		format,
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetMarklId(tipe, sig); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func Verify(
	publicKey, message, sig interfaces.MarklId,
) (err error) {
	defer errors.DeferredRecover(&err)

	switch publicKey.GetMarklFormat().GetMarklFormatId() {
	default:
		err = errors.Errorf(
			"not a public key: %q",
			publicKey.StringWithFormat(),
		)
		return

	case TypeIdEd25519Pub:
	}

	pub := ed25519.PublicKey(publicKey.GetBytes())

	if err = ed25519.VerifyWithOptions(
		pub,
		message.GetBytes(),
		sig.GetBytes(),
		&ed25519.Options{},
	); err != nil {
		err = errors.Err422UnprocessableEntity.Errorf(
			"invalid signature: %w. Signature: %q",
			err,
			sig.StringWithFormat(),
		)
		return
	}

	return
}
