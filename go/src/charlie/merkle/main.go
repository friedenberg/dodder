package merkle

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

type Generator interface {
	GeneratePrivateKey() (err error)
}

type (
	PublicKey = Data
	Sig       = Data
)

func NewKeyFromSeed(seed []byte) PrivateKey {
	return PrivateKey(ed25519.NewKeyFromSeed(seed))
}

func Sign(key PrivateKey, message []byte) (sig []byte, err error) {
	if sig, err = key.Sign(
		nil,
		message,
		&ed25519.Options{},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func VerifySignature(
	publicKey PublicKey,
	message []byte,
	sig []byte,
) (err error) {
	defer errors.DeferredRecover(&err)

	if err = ed25519.VerifyWithOptions(
		ed25519.PublicKey(publicKey),
		message,
		sig,
		&ed25519.Options{},
	); err != nil {
		err = errors.Err422UnprocessableEntity.Errorf(
			"invalid signature: %w. Signature: %x",
			err,
			sig,
		)
		return
	}

	return
}
