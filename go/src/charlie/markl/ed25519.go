package markl

import (
	"crypto"
	"crypto/ed25519"
	"io"

	"code.linenisgreat.com/dodder/go/src/alfa/errors"
)

// TODO convert into interface and structure like Hash
type PrivateKey ed25519.PrivateKey

func (privateKey PrivateKey) GetType() string {
	return "ed25519"
}

func (privateKey *PrivateKey) Generate(rand io.Reader) (err error) {
	var src ed25519.PrivateKey

	if _, src, err = ed25519.GenerateKey(rand); err != nil {
		err = errors.Wrap(err)
		return
	}

	*privateKey = PrivateKey(src)

	return
}

func (privateKey PrivateKey) Public() crypto.PublicKey {
	return ed25519.PrivateKey(privateKey).Public()
}

func (privateKey PrivateKey) Seed() []byte {
	return ed25519.PrivateKey(privateKey).Seed()
}

func (privateKey PrivateKey) Sign(
	rand io.Reader,
	message []byte,
	opts crypto.SignerOpts,
) (signature []byte, err error) {
	return ed25519.PrivateKey(privateKey).Sign(rand, message, opts)
}

func (privateKey *PrivateKey) UnmarshalBinary(data []byte) error {
	*privateKey = data
	return nil
}

func (privateKey PrivateKey) MarshalBinary() (data []byte, err error) {
	data = []byte(privateKey)
	return
}
