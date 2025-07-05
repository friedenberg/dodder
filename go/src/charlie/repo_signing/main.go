package repo_signing

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/bech32"
)

type Verifiable interface {
	GetPubKeyAndSig() (bech32.Value, bech32.Value, error)
}

type VerifiableFunc func() (bech32.Value, bech32.Value, error)

var _ = Verifiable(VerifiableFunc(nil))

func (funk VerifiableFunc) GetPubKeyAndSig() (bech32.Value, bech32.Value, error) {
	return funk()
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

func VerifySignatureOnVerifiable(
	verifiable Verifiable,
	message []byte,
) (err error) {
	var publicKey, sig bech32.Value

	if publicKey, sig, err = verifiable.GetPubKeyAndSig(); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = VerifySignature(publicKey.Data, message, sig.Data); err != nil {
		err = errors.Wrapf(err, "PubKey: %s, Sig: %s", publicKey, sig)
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
		err = errors.Wrapf(err, "invalid signature: %q", string(sig))
		return
	}

	return
}
