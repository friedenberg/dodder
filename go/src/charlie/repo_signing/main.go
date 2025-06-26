package repo_signing

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

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
	if err = ed25519.VerifyWithOptions(
		ed25519.PublicKey(publicKey),
		message,
		sig,
		&ed25519.Options{},
	); err != nil {
		err = errors.Wrapf(err, "invalid signature: %q", string(sig))
		return
	}

	return
}
