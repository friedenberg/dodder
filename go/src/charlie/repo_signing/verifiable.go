package repo_signing

import (
	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/bravo/blech32"
)

type (
	Verifiable interface {
		GetPubKeyAndSig() (blech32.Value, blech32.Value, error)
	}

	VerifiableFunc func() (blech32.Value, blech32.Value, error)
)

var _ = Verifiable(VerifiableFunc(nil))

func (funk VerifiableFunc) GetPubKeyAndSig() (blech32.Value, blech32.Value, error) {
	return funk()
}

func VerifySignatureOnVerifiable(
	verifiable Verifiable,
	message []byte,
) (err error) {
	var publicKey, sig blech32.Value

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
