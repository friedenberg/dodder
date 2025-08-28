package merkle

import (
	"crypto/ed25519"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
	"code.linenisgreat.com/dodder/go/src/alfa/interfaces"
	"code.linenisgreat.com/dodder/go/src/bravo/ui"
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

// TODO replace with Sign
func SignBytes(key PrivateKey, message []byte) (sig []byte, err error) {
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

// TODO replace with Verify
func VerifyBytes(
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

func Sign(
	key PrivateKey,
	message interfaces.BlobId,
	tipe string,
	dst interfaces.MutableBlobId,
) (err error) {
	var sig []byte

	if sig, err = key.Sign(
		nil,
		message.GetBytes(),
		&ed25519.Options{},
	); err != nil {
		err = errors.Wrap(err)
		return
	}

	if err = dst.SetMerkleId(tipe, sig); err != nil {
		err = errors.Wrap(err)
		return
	}

	return
}

func Verify(
	publicKey, message, sig interfaces.BlobId,
) (err error) {
	defer errors.DeferredRecover(&err)

	ui.Debug().Print(len(sig.GetBytes()))

	if err = ed25519.VerifyWithOptions(
		ed25519.PublicKey(publicKey.GetBytes()),
		message.GetBytes(),
		sig.GetBytes(),
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
